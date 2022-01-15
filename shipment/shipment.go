package shipment

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

// order ID muss ebenfalls gespeichert werden für Rückgabe, evtl. OrderID == ShipmentID
type Server struct {
	Nats       *nats.Conn
	Shipment   map[uint32]*api.NewShipmentRequest
	ShipmentID uint32
	api.UnimplementedShipmentServer
}

func (s Server) NewShipment(ctx context.Context, in *api.NewShipmentRequest) (*api.ShipmentReply, error) {
	log.Printf("received new shipment request of: articles: %v, availability: %v", in.GetArticles(), in.GetReady())

	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received new shipment request of: articles: %v, availability: %v", in.GetArticles(), in.GetReady())))
	if err != nil {
		panic(err)
	}

	var highestID uint32
	for key := range s.Shipment {
		if key > highestID {
			highestID = key
		}
	}
	s.ShipmentID = highestID + 1
	s.Shipment[s.ShipmentID] = in
	log.Printf("successfully created new shipment: id: %v, articles: %v, availability: %v", s.ShipmentID, in.GetArticles(), in.GetReady())
	err = s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("successfully created new shipment: id: %v, articles: %v, availability: %v", s.ShipmentID, in.GetArticles(), in.GetReady())))
	if err != nil {
		panic(err)
	}
	return &api.ShipmentReply{Id: s.ShipmentID, Articles: in.GetArticles(), Ready: in.GetReady()}, nil
}

//func (s Server) GetNeededArticles(ctx context.Context, in *api.ArticleRequest) (*api.ShipmentReply, error) {

//return &api.ShipmentReply{Id: in.GetId(), Articles: in.GetArticles(), Ready: in.GetReady()}, nil
//}

func (s Server) SendShipment(ctx context.Context, in *api.GetShipmentRequest) (*api.ShipmentReply, error) {
	log.Printf("received send shipment request: id: %v, articles: %v, availability: %v", in.GetId(), in.GetArticles(), in.GetReady())

	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received send shipment request: id: %v, articles: %v, availability: %v", in.GetId(), in.GetArticles(), in.GetReady())))
	if err != nil {
		panic(err)
	}
	// bei Stock nachfragen, ob Artikel vorhanden sind
	// map ready updaten
	out := in
	red := true
	for _, value := range in.GetReady() {
		if !value {
			red = value
		}
	}
	if !red {
		// warten??
	}
	return &api.ShipmentReply{Id: out.GetId(), Articles: out.GetArticles(), Ready: out.GetReady()}, nil
}

func (s Server) ReturnedShipment(ctx context.Context, in *api.GetShipmentRequest) (*api.ShipmentReply, error) {
	log.Printf("received return of shipment: id: %v, articles: %v, availability: %v", in.GetId(), in.GetArticles(), in.GetReady())

	err := s.Nats.Publish("log.shipment", []byte(fmt.Sprintf("received return of shipment: id: %v, articles: %v, availability: %v", in.GetId(), in.GetArticles(), in.GetReady())))
	if err != nil {
		panic(err)
	}
	// Artikel zurück zu Stock
	// payment zu Order
	// löschen?
	return &api.ShipmentReply{Id: in.GetId(), Articles: in.GetArticles(), Ready: in.GetReady()}, nil
}
