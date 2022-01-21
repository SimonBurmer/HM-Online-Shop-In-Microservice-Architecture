package stock

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats    *nats.EncodedConn
	Redis   *redis.Client
	Stock   map[uint32]*api.NewStockRequest
	StockID uint32
	api.UnimplementedStockServer
}

func (s *Server) AddStock(in *api.AddStockRequest) {
	log.Printf("received new article(s) with: ID: %v, quantity: %v", in.GetId(), in.GetAmount())

	err := s.Nats.Publish("log.stock", []byte(fmt.Sprintf("received new article(s) with: ID: %v, quantity: %v", in.GetId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}

	// article is already in DB
	amount := in.GetAmount()
	if val, ok := s.Stock[in.GetId()]; ok {

		// Send to Shipment
		if len(val.GetReserved()) > 0 {
			// delete könnte Fehler schmeißen -> out of range
			for key, value := range val.GetReserved() {
				if amount >= int32(value) {
					UpdateShipment := &api.ShipmentReadiness{Id: key, ArticleId: in.GetId(), Amount: value}
					err = s.Nats.Publish("shipment.articles", UpdateShipment)
					if err != nil {
						panic(err)
					}
					amount = amount - int32(value)
					delete(s.Stock[in.GetId()].GetReserved(), key)
				} else {
					value = uint32(in.GetAmount()) - value
					UpdateShipment := &api.ShipmentReadiness{Id: key, ArticleId: in.GetId(), Amount: uint32(amount)}
					err = s.Nats.Publish("shipment.articles", UpdateShipment)
					if err != nil {
						panic(err)
					}
					s.Stock[in.GetId()].Reserved[key] = value
					amount = 0
					break
				}
			}

		}
		// Rest in Stock
		s.Stock[in.GetId()].Amount = val.Amount + amount
	} else {
		// article is not in DB
		s.StockID = in.GetId()
		s.Stock[s.StockID] = &api.NewStockRequest{Amount: in.GetAmount()}

	}
	out := s.Stock[in.GetId()].Amount
	log.Printf("added new article(s) with: ID: %v, quantity: %v", in.GetId(), out)
	err = s.Nats.Publish("log.stock", []byte(fmt.Sprintf("added new article(s) with: ID: %v, quantity: %v", in.GetId(), out)))
	if err != nil {
		panic(err)
	}
}

func (s *Server) GetArticle(ctx context.Context, in *api.TakeArticle) (*api.GetReply, error) {
	log.Printf("received get article request of: id: %v, amount: %v, shipmentID: %v", in.GetId(), in.GetAmount(), in.GetShipmentId())

	err := s.Nats.Publish("log.stock", []byte(fmt.Sprintf("received get article request of: id: %v, amount: %v, shipmentID: %v", in.GetId(), in.GetAmount(), in.GetShipmentId())))
	if err != nil {
		panic(err)
	}

	out := s.Stock[in.GetId()]
	articleAmount := out.GetAmount() - in.GetAmount()

	if articleAmount < 0 {

		log.Printf("not enough stock available: id: %v, amount: %v", in.GetId(), in.GetAmount())
		err = s.Nats.Publish("log.stock", []byte(fmt.Sprintf("not enough stock available: id: %v, amount: %v", in.GetId(), in.GetAmount())))
		if err != nil {
			panic(err)
		}
		s.Stock[in.GetId()].Reserved[in.GetShipmentId()] = uint32(in.GetAmount()) - uint32(out.GetAmount())
		s.Stock[in.GetId()].Amount = 0

		// Bestellung der fehlenden Artikel beim Supplier
		neededAmount := articleAmount * (-1)
		NewOrderSupplier := &api.OrderArticleRequest{ArticleId: in.GetId(), Amount: uint32(neededAmount)}
		err = s.Nats.Publish("supplier.order", NewOrderSupplier)
		if err != nil {
			panic(err)
		}
		return &api.GetReply{Amount: out.GetAmount()}, nil

	}

	s.Stock[in.GetId()].Amount = articleAmount

	return &api.GetReply{Amount: in.GetAmount()}, nil
}

// eventuell Amount zurück geben
func (s *Server) GetStock(ctx context.Context, in *api.ArticleID) (*api.GetStockReply, error) {
	log.Printf("received get stock request with: id: %v", in.GetId())

	err := s.Nats.Publish("log.stock", []byte(fmt.Sprintf("received get stock request with: id: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	answer := true
	_, ok := s.Stock[in.GetId()]
	if !ok {
		err = s.Nats.Publish("log.stock", []byte(fmt.Sprintf("no article with Id: %v", in.GetId())))
		if err != nil {
			panic(err)
		}
		log.Fatalf("no article with Id: %v", in.GetId())
	}
	if s.Stock[in.GetId()].GetAmount() <= 0 {
		answer = false
	}
	return &api.GetStockReply{Answer: answer}, nil
}

func (s *Server) CancelReserved(in *api.CancelReservedRequest) {
	delete(s.Stock[in.GetId()].GetReserved(), in.GetShipmentId())
}
