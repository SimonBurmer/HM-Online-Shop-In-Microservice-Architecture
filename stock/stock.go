package stock

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats    *nats.Conn
	Stock   map[int32]int32
	StockID int32
	api.UnimplementedStockServer
}

func (s Server) AddStock(ctx context.Context, in *api.NewStock) (*api.GetReply, error) {
	log.Printf("received new article(s) with: ID: %v, quantity: %v", in.GetId(), in.GetAmount())

	err := s.Nats.Publish("log.stock", []byte(fmt.Sprintf("received new article(s) with: ID: %v, quantity: %v", in.GetId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}

	// article is already in DB
	if val, ok := s.Stock[in.GetId()]; ok {
		s.Stock[in.GetId()] = val + in.GetAmount()
	} else {
		// article is not in DB
		s.StockID++
		s.Stock[s.StockID] = in.GetAmount()
	}

	out := s.Stock[in.GetId()]
	return &api.GetReply{Amount: out}, nil
}

func (s Server) GetArticle(ctx context.Context, in *api.TakeArticle) (*api.GetReply, error) {
	log.Printf("received get article request of: id: %v, amount: %v", in.GetId(), in.GetAmount())

	err := s.Nats.Publish("log.stock", []byte(fmt.Sprintf("received get article request of: id: %v, amount: %v", in.GetId(), in.GetAmount())))
	if err != nil {
		panic(err)
	}

	out := s.Stock[in.GetId()]
	s.Stock[in.GetId()] = out - in.GetAmount()
	out = out - in.GetAmount() // TODO Fehlerbehandlung und negativen Stock

	return &api.GetReply{Amount: out}, nil
}

func (s Server) GetStock(ctx context.Context, in *api.ArticleID) (*api.GetStockReply, error) {
	log.Printf("received get stock request with: id: %v", in.GetId())

	err := s.Nats.Publish("log.stock", []byte(fmt.Sprintf("received get stock request with: id: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	answer := false
	if _, ok := s.Stock[in.GetId()]; ok { // article with id exists
		answer = true
	}
	return &api.GetStockReply{Answer: answer}, nil
}
