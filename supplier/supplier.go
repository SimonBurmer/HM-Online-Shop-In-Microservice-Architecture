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
	Supplier   map[uint32]*api.NewArticles
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

	return &api.GetSupplierReply{Answer: "ok"}, nil
}

// asynchron aufgerufene Funktion
func (s *Server) OrderSupplies(in *api.OrderArticleRequest) {
	log.Printf("received order for article(s) with: Order ID: %v, ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount())
	err := s.Nats.Publish("log.supplier", []byte(fmt.Sprintf("received order for article(s) with: Order ID: %v, ID: %v, quantity: %v", in.GetOrderId(), in.GetArticleId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}

	s.SupplierID++
	s.Supplier[s.SupplierID].Amount = in.GetAmount()
	s.Supplier[s.SupplierID].ArticleId = in.GetArticleId()
	// TODO
	//s.Supplier[s.SupplierID].NameSupplier = s.OrderArticle(context.TODO(), in);

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
