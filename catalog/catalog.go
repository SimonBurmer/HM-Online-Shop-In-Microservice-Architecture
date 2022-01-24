package catalog

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type Server struct {
	Nats      *nats.EncodedConn
	Redis     *redis.Client
	Catalog   map[uint32]*api.NewCatalog
	CatalogID uint32
	api.UnimplementedCatalogServer
}

func (s *Server) GetAvailability(in *api.GetCatalog) (answer bool) {
	log.Printf("availability check: id:%v", in.GetId())
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
	stock_r, stock_err := stock.GetStock(stock_ctx, &api.ArticleID{Id: in.GetId()})
	if stock_err != nil {
		st, ok := status.FromError(stock_err)
		if !ok {
			log.Fatalf("An unexpected error occurred: %v", stock_err)
		}
		log.Printf("GetStock failed: %v", st.Message())
	}
	log.Printf("Article available in Stock: Id: %v, Available: %v", in.GetId(), stock_r.GetAnswer())

	return stock_r.GetAnswer()

}

func (s *Server) GetCatalogInfo(ctx context.Context, in *api.GetCatalog) (*api.CatalogReplyInfo, error) {
	log.Printf("received request for the article with: id: %v", in.GetId())

	err := s.Nats.Publish("log.catalog", api.Log{Message: fmt.Sprintf("received request for the article with: id: %v", in.GetId()), Subject: "Catalog.GetCatalogInfo"})
	if err != nil {
		panic(err)
	}

	out, ok := s.Catalog[in.GetId()]
	if !ok {
		err = s.Nats.Publish("log.catalog", api.Log{Message: fmt.Sprintf("no article with Id: %v", in.GetId()), Subject: "Catalog.GetCatalogInfo"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("no article with Id: %v", in.GetId())
	}

	log.Printf("successfully loaded catalog of: id: %v, name: %v, description: %v, price: %v", in.GetId(), out.GetName(), out.GetDescription(), out.GetPrice())
	err = s.Nats.Publish("log.catalog", api.Log{Message: fmt.Sprintf("successfully loaded catalog of: id: %v, name: %v, description: %v, price: %v", in.GetId(), out.GetName(), out.GetDescription(), out.GetPrice()), Subject: "Catalog.GetCatalogInfo"})
	if err != nil {
		panic(err)
	}

	available := s.GetAvailability(in)

	return &api.CatalogReplyInfo{Id: s.CatalogID, Name: out.GetName(), Description: out.GetDescription(), Price: out.GetPrice(), Availability: available}, nil
}

func (s *Server) NewCatalogArticle(ctx context.Context, in *api.NewCatalog) (*api.CatalogReply, error) {
	log.Printf("received new catalog request of: name: %v, description: %v, price: %v", in.GetName(), in.GetDescription(), in.GetPrice())

	err := s.Nats.Publish("log.catalog", api.Log{Message: fmt.Sprintf("received new catalog request of: name: %v, description: %v, price: %v", in.GetName(), in.GetDescription(), in.GetPrice()), Subject: "Catalog.NewCatalogArticle"})
	if err != nil {
		panic(err)
	}

	s.CatalogID = s.CatalogID + 1
	s.Catalog[s.CatalogID] = in
	log.Printf("successfully created new catalog article: id: %v, name: %v, description: %v, price: %v", s.CatalogID, in.GetName(), in.GetDescription(), in.GetPrice())
	err = s.Nats.Publish("log.catalog", api.Log{Message: fmt.Sprintf("successfully created new catalog article: id: %v, name: %v, description: %v, price: %v", s.CatalogID, in.GetName(), in.GetDescription(), in.GetPrice()), Subject: "Catalog.NewCatalogArticle"})
	if err != nil {
		panic(err)
	}

	newStockEntry := &api.AddStockRequest{Id: s.CatalogID, Amount: 0}
	err = s.Nats.Publish("stock.add", newStockEntry)
	if err != nil {
		panic(err)
	}
	log.Printf("updated stock: Id:%v", newStockEntry.GetId())

	return &api.CatalogReply{Id: s.CatalogID, Name: in.GetName(), Description: in.GetDescription(), Price: in.GetPrice()}, nil
}

func (s *Server) UpdateCatalog(ctx context.Context, in *api.UpdatedData) (*api.CatalogReply, error) {
	log.Printf("received update catalog request with: id: %v, name: %v, description: %v, price: %v", in.GetId(), in.GetName(), in.GetDescription(), in.GetPrice())

	err := s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("received update catalog request with: id: %v, name: %v, description: %v, price: %v", in.GetId(), in.GetName(), in.GetDescription(), in.GetPrice())))
	if err != nil {
		panic(err)
	}

	tmp := &api.NewCatalog{Name: in.GetName(), Description: in.GetDescription(), Price: in.GetPrice()}
	s.Catalog[in.GetId()] = tmp
	log.Printf("successfully updated catalog: id: %v, name: %v, description: %v, price: %v", s.CatalogID, in.GetName(), in.GetDescription(), in.GetPrice())
	err = s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("successfully updated catalog: id: %v, name: %v, description: %v, price: %v", s.CatalogID, in.GetName(), in.GetDescription(), in.GetPrice())))
	if err != nil {
		panic(err)
	}
	return &api.CatalogReply{Id: s.CatalogID, Name: in.GetName(), Description: in.GetDescription(), Price: in.GetPrice()}, nil
}

func (s *Server) DeleteCatalog(ctx context.Context, in *api.GetCatalog) (*api.CatalogReply, error) {
	log.Printf("received delete catalog request of: id: %v", in.GetId())

	err := s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("received delete catalog request of: id: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	out := s.Catalog[in.GetId()]
	delete(s.Catalog, in.GetId())
	log.Printf("successfully deleted catalog article: id: %v, name: %v, description: %v, price: %v", in.GetId(), out.GetName(), out.GetDescription(), out.GetPrice())
	err = s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("successfully deleted catalog article: id: %v, name: %v, description: %v, price: %v", in.GetId(), out.GetName(), out.GetDescription(), out.GetPrice())))
	if err != nil {
		panic(err)
	}
	return &api.CatalogReply{Id: in.GetId(), Name: out.GetName(), Description: out.GetDescription(), Price: out.GetPrice()}, nil
}

func (s *Server) First() {
	s.Catalog[1] = &api.NewCatalog{Name: "Printer123", Description: "Very good printer!", Price: 102.50}
	newStockEntry := &api.AddStockRequest{Id: 1, Amount: 3}
	err := s.Nats.Publish("stock.add", newStockEntry)
	if err != nil {
		panic(err)
	}
	log.Printf("filling stock: Id:%v", newStockEntry.GetId())

	s.Catalog[2] = &api.NewCatalog{Name: "Lamp", Description: "Ligths up the room", Price: 42.00}
	newStockEntry = &api.AddStockRequest{Id: 2, Amount: 2}
	err = s.Nats.Publish("stock.add", newStockEntry)
	if err != nil {
		panic(err)
	}
	log.Printf("filling stock: Id:%v", newStockEntry.GetId())

	s.Catalog[3] = &api.NewCatalog{Name: "Microwave", Description: "Newest technology. Heats food up 10 times faster", Price: 63.99}
	newStockEntry = &api.AddStockRequest{Id: 3, Amount: 0}
	err = s.Nats.Publish("stock.add", newStockEntry)
	if err != nil {
		panic(err)
	}
	log.Printf("filling stock: Id:%v", newStockEntry.GetId())

	s.Catalog[4] = &api.NewCatalog{Name: "Toaster", Description: "For the perfect toast", Price: 19.99}
	newStockEntry = &api.AddStockRequest{Id: 4, Amount: 0}
	err = s.Nats.Publish("stock.add", newStockEntry)
	if err != nil {
		panic(err)
	}
	log.Printf("filling stock: Id:%v", newStockEntry.GetId())

}
