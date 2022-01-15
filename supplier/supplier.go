package supplier

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats      *nats.Conn
	Supplie   map[uint32]uint32
	SupplieID uint32
	api.UnimplementedSupplierServer
}

// orders articles from the Supplier API
func (s Server) orderArticle(in *api.NewArticles) (string, error) {

	// TODO
	return "done", nil
}

func (s Server) DeliveredArticle(ctx context.Context, in *api.NewArticles) (*api.GetSupplierReply, error) {
	log.Printf("received new article(s) with: ID: %v, quantity: %v", in.GetId(), in.GetAmount())

	err := s.Nats.Publish("log.supplier", []byte(fmt.Sprintf("received new article(s) with: ID: %v, quantity: %v", in.GetId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}
	return &api.GetSupplierReply{Answer: "ok"}, nil
}
