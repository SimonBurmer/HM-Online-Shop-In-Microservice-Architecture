package customer

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats       *nats.Conn
	Customers  map[uint32]*api.NewCustomerRequest
	CustomerID uint32
	api.UnimplementedCustomerServer
}

func (s Server) NewCustomer(ctx context.Context, in *api.NewCustomerRequest) (*api.CustomerReply, error) {
	log.Printf("received new customer request of: name: %v, address: %v", in.GetName(), in.GetAddress())

	// Indirekte Kommunikation über NATS (Stellt Logger einer Nachricht rein)
	err := s.Nats.Publish("log.customer", []byte(fmt.Sprintf("received new customer of: name: %v, address: %v", in.GetName(), in.GetAddress())))
	if err != nil {
		panic(err)
	}

	s.CustomerID++
	s.Customers[s.CustomerID] = in
	log.Printf("successfully created new customer: name: %v, address: %v", in.GetName(), in.GetAddress())

	return &api.CustomerReply{Id: s.CustomerID, Name: in.GetName(), Address: in.GetAddress()}, nil
}

func (s Server) GetCustomer(ctx context.Context, in *api.GetCustomerRequest) (*api.CustomerReply, error) {
	log.Printf("received get customer request of: id: %v", in.GetId())

	err := s.Nats.Publish("log.customer", []byte(fmt.Sprintf("received new customer request of: id: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	out := s.Customers[in.GetId()]
	// Fehlerbehandlung!!
	//log.Fatalf("No Customer with id: %v", in.GetId())

	log.Printf("successfully loaded customer of: id: %v", in.GetId())
	return &api.CustomerReply{Id: s.CustomerID, Name: out.GetName(), Address: out.GetAddress()}, nil
}
