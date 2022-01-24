package shipment

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
	Nats       *nats.EncodedConn
	Redis      *redis.Client
	Shipment   map[uint32]*api.ShipmentStorage
	ShipmentID uint32
	api.UnimplementedShipmentServer
}

func (s *Server) NewShipment(in *api.NewShipmentRequest) {
	log.Printf("received new shipment request of: order ID: %v articles: %v, address: %v", in.GetOrderID(), in.GetArticles(), in.GetAddress())
	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received new shipment request of: order ID: %v, articles: %v, address: %v", in.GetOrderID(), in.GetArticles(), in.GetAddress())))
	if err != nil {
		panic(err)
	}

	s.ShipmentID = in.GetOrderID()
	m := make(map[uint32]uint32)
	for key := range in.GetArticles() {
		m[key] = 0
	}
	s.Shipment[s.ShipmentID] = &api.ShipmentStorage{Address: in.GetAddress(), Articles: in.GetArticles(), Ready: m}
	log.Printf("successfully created new shipment: id: %v, order ID: %v, articles: %v, availability: %v, address: %v", s.ShipmentID, in.GetOrderID(), in.GetArticles(), s.Shipment[s.ShipmentID].GetReady(), in.GetAddress())
	err = s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("successfully created new shipment: id: %v, order ID: %v articles: %v, availability: %v, address: %v", s.ShipmentID, in.GetOrderID(), in.GetArticles(), s.Shipment[s.ShipmentID].GetReady(), in.GetAddress())))
	if err != nil {
		panic(err)
	}
	for index, element := range in.GetArticles() {
		// Artikel aus Stock ausbuchen
		log.Printf("Get articles from Stock: id:%v, amount:%v", index, element)
		stock := s.getConnectionStock(&api.ShipmentReadiness{Id: s.ShipmentID, ArticleId: index, Amount: element})
		s.Shipment[s.ShipmentID].Ready[index] = uint32(stock.GetAmount())
	}
	// Überprüfen ob Shipment schon versendet werden kann
	ready := &api.ShipmentReadiness{Id: s.ShipmentID}
	s.ShipmentReady(ready)

}

func (s *Server) ShipmentReady(in *api.ShipmentReadiness) {
	log.Printf("received shipment ready request of: order ID: %v article ID: %v, amount: %v", in.GetId(), in.GetArticleId(), in.GetAmount())
	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received shipment ready request of: order ID: %v article ID: %v, amount: %v", in.GetId(), in.GetArticleId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}
	have := s.Shipment[in.GetId()].Ready[in.GetArticleId()]
	s.Shipment[in.GetId()].Ready[in.GetArticleId()] = have + in.GetAmount()
	articlesAvailable := s.Shipment[in.GetId()].GetReady()
	ready := true
	for key, element := range s.Shipment[s.ShipmentID].GetArticles() {
		if element != articlesAvailable[key] {
			ready = false
		}
	}
	if ready {
		log.Printf("Shipment %v is ready to be send", in.GetId())
		s.SendShipment(context.TODO(), &api.GetShipmentRequest{Id: in.GetId(), Articles: s.Shipment[s.ShipmentID].GetArticles(), Ready: s.Shipment[s.ShipmentID].GetReady()})
	}
}

func (s *Server) SendShipment(ctx context.Context, in *api.GetShipmentRequest) (*api.ShipmentReply, error) {
	log.Printf("received send shipment request: id: %v, articles: %v, availability: %v", in.GetId(), in.GetArticles(), in.GetReady())

	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received send shipment request: id: %v, articles: %v, availability: %v", in.GetId(), in.GetArticles(), in.GetReady())))
	if err != nil {
		panic(err)
	}

	informOrder := &api.OrderShipmentUpdate{OrderId: in.GetId()}
	err = s.Nats.Publish("order.shipment", informOrder)
	if err != nil {
		panic(err)
	}
	address := s.Shipment[in.GetId()].GetAddress()

	// Hier würde die Anbindung an die API erfolgen
	log.Printf("sucessfully send shipment: id: %v, articles: %v, address: %v", in.GetId(), in.GetArticles(), s.Shipment[s.ShipmentID].GetAddress())

	err = s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("sucessfully send shipment: id: %v, articles: %v, address: %v", in.GetId(), in.GetArticles(), s.Shipment[s.ShipmentID].GetAddress())))
	if err != nil {
		panic(err)
	}
	return &api.ShipmentReply{Id: in.GetId(), Articles: in.GetArticles(), Ready: in.GetReady(), Address: address}, nil
}

func (s *Server) CancelShipment(in *api.OrderID) {
	log.Printf("received cancellation of shipment of: ID: %v", in.GetId())
	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received cancellation of shipment of: ID: %v", in.GetId())))
	if err != nil {
		panic(err)
	}
	out := s.Shipment[in.GetId()]
	for key, element := range out.GetArticles() {
		// Reservierung von nicht vorhandenen Artikeln löschen
		cancelReserved := &api.CancelReservedRequest{ShipmentId: in.GetId(), Id: key}
		err := s.Nats.Publish("stock.cancel", cancelReserved)
		if err != nil {
			panic(err)
		}
		log.Printf("Request to cancel reservation send to Stock: Id: %v, Amount: %v", key, element)

		// Schon ausgebuchte Artikel an Stock zurückschicken
		addedStock := &api.AddStockRequest{Id: key, Amount: int32(element)}
		err = s.Nats.Publish("stock.add", addedStock)
		if err != nil {
			panic(err)
		}
		log.Printf("Articles send to Stock: Id: %v, Amount: %v", key, element)

	}

}

func (s *Server) ReturnDefectArticle(ctx context.Context, in *api.ShipmentReturnRequest) (*api.ReturnReply, error) {

	log.Printf("received return of defect article: ID: %v, article ID: %v, amount: %v", in.GetId(), in.GetArticleId(), in.GetAmount())
	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received return of defect article: ID: %v, article ID: %v, amount: %v", in.GetId(), in.GetArticleId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}

	s.Shipment[in.GetId()].Ready[in.GetArticleId()] = s.Shipment[in.GetId()].Articles[in.GetArticleId()] - in.GetAmount()
	stock := s.getConnectionStock(&api.ShipmentReadiness{Id: in.GetId(), ArticleId: in.GetArticleId(), Amount: in.GetAmount()})
	s.Shipment[s.ShipmentID].Ready[in.GetArticleId()] = uint32(stock.GetAmount())
	ready := &api.ShipmentReadiness{Id: s.ShipmentID}
	s.ShipmentReady(ready)
	return &api.ReturnReply{Id: in.GetId(), ArticleId: in.GetArticleId(), Amount: in.GetAmount()}, nil
}

func (s *Server) Refund(ctx context.Context, in *api.ShipmentReturnRequest) (*api.ReturnReply, error) {
	log.Printf("received refund request of: order ID: %v article ID: %v, amount: %v", in.GetId(), in.GetArticleId(), in.GetAmount())
	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received refund request of: order ID: %v article ID: %v, amount: %v", in.GetId(), in.GetArticleId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}
	// order Bescheid geben
	refundOrder := &api.RefundArticleRequest{OrderId: in.GetId(), ArticleId: in.GetArticleId()}
	err = s.Nats.Publish("order.refund", refundOrder)
	if err != nil {
		panic(err)
	}
	log.Printf("Sent refund request to order")

	// artikel löschen??
	return &api.ReturnReply{Id: in.GetId(), ArticleId: in.GetArticleId(), Amount: in.GetAmount()}, nil
}

func (s *Server) getConnectionStock(in *api.ShipmentReadiness) *api.GetReply {
	redisVal := s.Redis.Get(context.TODO(), "stock")
	if redisVal == nil {
		log.Fatalf("service not registered")
	}
	address, err := redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	stock_con, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	stock := api.NewStockClient(stock_con)
	stock_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer stock_con.Close()
	defer cancel()

	// - Stock-Informationen holen
	stock_r, stock_err := stock.GetArticle(stock_ctx, &api.TakeArticle{Id: in.GetArticleId(), Amount: int32(in.GetAmount()), ShipmentId: in.GetId()})
	if stock_err != nil {
		log.Fatalf("could not get articles of: articleId: %v", in.GetArticleId())
	}

	return stock_r
}
