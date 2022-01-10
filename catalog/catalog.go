package catalog

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats      *nats.Conn
	Redis     *redis.Client
	Catalog   map[uint32]*api.NewCatalog
	CatalogID uint32
	api.UnimplementedCatalogServer
}

func (s Server) GetCatalogInfo(ctx context.Context, in *api.GetCatalog) (*api.CatalogReply, error) {
	log.Printf("received request for the article with: id: %v", in.GetId())

	err := s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("received request for the article with: id: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	out, ok := s.Catalog[in.GetId()]
	if !ok {
		err = s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("no article with Id: %v", in.GetId())))
		if err != nil {
			panic(err)
		}
		log.Fatalf("no article with Id: %v", in.GetId())
	}

	log.Printf("successfully loaded catalog of: id: %v, name: %v, description: %v, price: %v", in.GetId(), out.GetName(), out.GetDescription(), out.GetPrice())
	err = s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("successfully loaded catalog of: id: %v, name: %v, description: %v, price: %v", in.GetId(), out.GetName(), out.GetDescription(), out.GetPrice())))
	if err != nil {
		panic(err)
	}

	return &api.CatalogReply{Id: s.CatalogID, Name: out.GetName(), Description: out.GetDescription(), Price: out.GetPrice()}, nil
}

func (s Server) NewCatalogArticle(ctx context.Context, in *api.NewCatalog) (*api.CatalogReply, error) {
	log.Printf("received new catalog request of: name: %v, description: %v, price: %v", in.GetName(), in.GetDescription(), in.GetPrice())

	err := s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("received new catalog request of: name: %v, description: %v, price: %v", in.GetName(), in.GetDescription(), in.GetPrice())))
	if err != nil {
		panic(err)
	}

	s.CatalogID++
	s.Catalog[s.CatalogID] = in
	log.Printf("successfully created new catalog article: id: %v, name: %v, description: %v, price: %v", s.CatalogID, in.GetName(), in.GetDescription(), in.GetPrice())
	err = s.Nats.Publish("log.catalog", []byte(fmt.Sprintf("successfully created new catalog article: id: %v, name: %v, description: %v, price: %v", s.CatalogID, in.GetName(), in.GetDescription(), in.GetPrice())))
	if err != nil {
		panic(err)
	}
	return &api.CatalogReply{Id: s.CatalogID, Name: in.GetName(), Description: in.GetDescription(), Price: in.GetPrice()}, nil
}

func (s Server) UpdateCatalog(ctx context.Context, in *api.UpdatedData) (*api.CatalogReply, error) {
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

func (s Server) DeleteCatalog(ctx context.Context, in *api.GetCatalog) (*api.CatalogReply, error) {
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
