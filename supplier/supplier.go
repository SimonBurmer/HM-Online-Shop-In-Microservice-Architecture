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

	log.Printf("received order for ordering article(s) with: Order ID: %v, ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount())
	err := s.Nats.Publish("log.supplier", []byte(fmt.Sprintf("received order for ordering article(s) with: Order ID: %v, ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}
	// TODO
	// return name supplier??
	return &api.SupplierName{Name: "yes"}, nil
}

func (s *Server) DeliveredArticle(ctx context.Context, in *api.NewArticles) (*api.GetSupplierReply, error) {
	log.Printf("received new article(s) with: Order ID: %v, ID: %v, quantity: %v, Supplier name: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount(), in.GetNameSupplier())

	err := s.Nats.Publish("log.supplier", []byte(fmt.Sprintf("received new article(s) with: Order ID: %v, ID: %v, quantity: %v, Supplier name: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount(), in.GetNameSupplier())))
	if err != nil {
		panic(err)
	}

	addedStock := &api.AddStockRequest{Id: in.GetArticleId(), Amount: int32(in.GetAmount())}
	err = s.Nats.Publish("stock.add", addedStock)
	if err != nil {
		panic(err)
	}
	log.Printf("Articles send to Stock: Id: %v, Amount: %v", in.GetArticleId(), in.GetAmount())

	return &api.GetSupplierReply{OrderId: in.GetOrderId(), ArticleId: in.GetArticleId(), Amount: in.GetAmount(), NameSupplier: in.GetNameSupplier()}, nil
}

// asynchron aufgerufene Funktion
func (s *Server) OrderSupplies(in *api.OrderArticleRequest) {
	log.Printf("received order for article(s) with: Order ID: %v, ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount())
	err := s.Nats.Publish("log.supplier", []byte(fmt.Sprintf("received order for article(s) with: Order ID: %v, ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}

	s.SupplierID = s.SupplierID + 1
	log.Printf("new supplier ID")
	s.Supplier[s.SupplierID] = &api.SupplierStorage{ArticleId: in.GetArticleId(), Amount: in.GetAmount()}
	log.Printf("Article Id has been stored: %v", s.Supplier[s.SupplierID].GetArticleId())

	// TODO
	name, _ := s.OrderArticle(context.TODO(), in)
	s.Supplier[s.SupplierID] = &api.SupplierStorage{NameSupplier: name.GetName()}
	log.Printf("Sucessfully ordered article(s) with: id: %v, amount: %v, name supplier: %v", in.GetArticleId(), in.GetAmount(), name.GetName())

}

/*
stock_redisVal := s.Redis.Get(context.TODO(), "stock")
	if stock_redisVal == nil {
		log.Fatal("service not registered")
	}
	stock_address, err := stock_redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	stock_conn, err := grpc.Dial(stock_address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer stock_conn.Close()
	stock := api.NewStockClient(stock_conn)
	stock_ctx, stock_cancel := context.WithTimeout(context.Background(), time.Second)
	defer stock_cancel()

	stock_r, stock_err := stock.AddStock(stock_ctx, &api.AddStockRequest{Id: in.GetArticleId(), Amount: int32(in.GetAmount())})
	if stock_err != nil {
		log.Fatalf("Direct communication with stock failed: %v", stock_r)
	}
*/
