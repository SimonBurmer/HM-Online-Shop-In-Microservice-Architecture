package supplier

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats       *nats.EncodedConn
	Redis      *redis.Client
	Supplier   map[uint32]*api.SupplierStorage
	SupplierID uint32
	api.UnimplementedSupplierServer
}

// orders articles from the Supplier API
func (s *Server) OrderArticle(ctx context.Context, in *api.OrderArticleRequest) (*api.SupplierName, error) {

	log.Printf("received order for ordering article(s) from outside suppliers: Order ID: %v, ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount())
	err := s.Nats.Publish("log.supplier", api.Log{Message: fmt.Sprintf("received order for ordering article(s) from outside suppliers: Order ID: %v, ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount()), Subject: "Supplier.OrderArticle"})
	if err != nil {
		panic(err)
	}
	// return name  of standard supplier bc there's no API
	name := "Amazon"
	log.Printf("Sucessfully ordered article(s) with: id: %v, amount: %v, name supplier: %v", in.GetArticleId(), in.GetAmount(), name)
	return &api.SupplierName{Name: name}, nil
}

func (s *Server) DeliveredArticle(ctx context.Context, in *api.NewArticles) (*api.GetSupplierReply, error) {
	log.Printf("received new article(s) with: Order ID: %v, ID: %v, quantity: %v, Supplier name: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount(), in.GetNameSupplier())

	err := s.Nats.Publish("log.supplier", api.Log{Message: fmt.Sprintf("received new article(s) with: Order ID: %v, ID: %v, quantity: %v, Supplier name: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount(), in.GetNameSupplier()), Subject: "Supplier.DeliveredArticle"})
	if err != nil {
		panic(err)
	}

	addedStock := &api.AddStockRequest{Id: in.GetArticleId(), Amount: int32(in.GetAmount())}
	err = s.Nats.Publish("stock.add", addedStock)
	if err != nil {
		panic(err)
	}
	log.Printf("Articles send to Stock: Id: %v, Amount: %v", in.GetArticleId(), in.GetAmount())
	err = s.Nats.Publish("log.supplier", api.Log{Message: fmt.Sprintf("Articles send to Stock: Id: %v, Amount: %v", in.GetArticleId(), in.GetAmount())})
	if err != nil {
		panic(err)
	}
	return &api.GetSupplierReply{OrderId: in.GetOrderId(), ArticleId: in.GetArticleId(), Amount: in.GetAmount(), NameSupplier: in.GetNameSupplier()}, nil
}

// asynchron aufgerufene Funktion
func (s *Server) OrderSupplies(in *api.OrderArticleRequest) {
	log.Printf("received order for article(s) with: order ID: %v, article ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount())
	err := s.Nats.Publish("log.supplier", api.Log{Message: fmt.Sprintf("received order for article(s) with: order ID: %v, article ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount()), Subject: "Supplier.OrderSupplies"})
	if err != nil {
		panic(err)
	}

	s.SupplierID = s.SupplierID + 1
	s.Supplier[s.SupplierID] = &api.SupplierStorage{ArticleId: in.GetArticleId(), Amount: in.GetAmount()}
	log.Printf("Sucessfully created new supplier order: supplierId:%v, articleId:%v, amount:%v", s.SupplierID, in.GetArticleId(), in.GetAmount())

	name, _ := s.OrderArticle(context.TODO(), in)
	s.Supplier[s.SupplierID] = &api.SupplierStorage{NameSupplier: name.GetName()}
}
