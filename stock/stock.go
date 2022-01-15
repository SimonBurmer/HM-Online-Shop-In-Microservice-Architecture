package stock

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"google.golang.org/grpc"
)

type Server struct {
	Nats    *nats.Conn
	Redis   *redis.Client
	Stock   map[uint32]*api.NewStockRequest
	StockID uint32
	api.UnimplementedStockServer
}

func (s *Server) AddStock(ctx context.Context, in *api.AddStockRequest) (*api.GetReply, error) {
	log.Printf("received new article(s) with: ID: %v, quantity: %v", in.GetId(), in.GetAmount())

	err := s.Nats.Publish("log.stock", []byte(fmt.Sprintf("received new article(s) with: ID: %v, quantity: %v", in.GetId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}

	// article is already in DB
	if val, ok := s.Stock[in.GetId()]; ok {
		if val.GetReserved() > 0 {
			// Send to Shipment
			// Rest in Stock
		}
		s.Stock[in.GetId()].Amount = val.Amount + in.GetAmount()
	} else {
		// article is not in DB
		s.StockID = in.GetId()
		s.Stock[s.StockID] = &api.NewStockRequest{Amount: in.GetAmount(), Reserved: 0}

	}
	out := s.Stock[in.GetId()].Amount
	log.Printf("added new article(s) with: ID: %v, quantity: %v", in.GetId(), out)
	err = s.Nats.Publish("log.stock", []byte(fmt.Sprintf("added new article(s) with: ID: %v, quantity: %v", in.GetId(), out)))
	if err != nil {
		panic(err)
	}

	return &api.GetReply{Amount: out}, nil
}

func (s *Server) GetArticle(ctx context.Context, in *api.TakeArticle) (*api.GetReply, error) {
	log.Printf("received get article request of: id: %v, amount: %v", in.GetId(), in.GetAmount())

	err := s.Nats.Publish("log.stock", []byte(fmt.Sprintf("received get article request of: id: %v, amount: %v", in.GetId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}

	out := s.Stock[in.GetId()]
	s.Stock[in.GetId()].Amount = out.Amount - in.GetAmount()
	if out.Amount-in.GetAmount() < 0 {

		log.Printf("not enough stock available: id: %v, amount: %v", in.GetId(), in.GetAmount())
		err = s.Nats.Publish("log.stock", []byte(fmt.Sprintf("not enough stock available: id: %v, amount: %v", in.GetId(), in.GetAmount())))
		if err != nil {
			panic(err)
		}
		s.Stock[in.GetId()].Reserved = uint32(in.GetAmount()) - uint32(out.GetAmount())

		// Bestellung bei Supplier aufgeben
		// Mithilfe von Redis Verbindung zu Supplier aufbauen
		supplier_redisVal := s.Redis.Get(context.TODO(), "supplier")
		if supplier_redisVal == nil {
			log.Fatal("service not registered")
		}
		supplier_address, err := supplier_redisVal.Result()
		if err != nil {
			log.Fatalf("error while trying to get the result %v", err)
		}
		supplier_conn, err := grpc.Dial(supplier_address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer supplier_conn.Close()
		supplier := api.NewPaymentClient(supplier_conn)
		supplier_ctx, supplier_cancel := context.WithTimeout(context.Background(), time.Second)
		defer supplier_cancel()

		// Kommunikation mit Supplier:
		// - Neue Nachbestellung von Artikeln
		//TODO Methode in Supplier erstellen
		supplier_r, supplier_err := supplier.NewPayment(supplier_ctx, &api.NewPaymentRequest{OrderId: 1, Value: 33.98})
		if supplier_err != nil {
			log.Fatalf("direct communication with supplier failed: %v", supplier_r)
		}
		log.Printf("reordered article: Id:%v, OrderId:%v, Value:%v", supplier_r.GetId(), supplier_r.GetOrderId(), supplier_r.GetValue())

	}
	out.Amount = out.Amount - in.GetAmount()

	return &api.GetReply{Amount: out.GetAmount()}, nil
}

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
	return &api.GetStockReply{Answer: answer}, nil
}
