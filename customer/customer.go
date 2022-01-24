package customer

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	Nats       *nats.EncodedConn
	Customers  map[uint32]*api.NewCustomerRequest
	CustomerID uint32
	api.UnimplementedCustomerServer
}

func (s *Server) NewCustomer(ctx context.Context, in *api.NewCustomerRequest) (*api.CustomerReply, error) {
	log.Printf("received new customer request of: name: %v, address: %v", in.GetName(), in.GetAddress())

	// Indirekte Kommunikation über NATS (Stellt Logger einer Nachricht rein)
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received new customer request of: name: %v, address: %v", in.GetName(), in.GetAddress()), Subject: "Customer.NewCustomer"})
	if err != nil {
		panic(err)
	}

	s.CustomerID = s.CustomerID + 1
	s.Customers[s.CustomerID] = in
	log.Printf("successfully created new customer: id: %v, name: %v, address: %v", s.CustomerID, in.GetName(), in.GetAddress())

	return &api.CustomerReply{Id: s.CustomerID, Name: in.GetName(), Address: in.GetAddress()}, nil
}

func (s *Server) GetCustomer(ctx context.Context, in *api.GetCustomerRequest) (*api.CustomerReply, error) {
	log.Printf("received get customer request of: id: %v", in.GetId())

	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received get customer request of: id: %v", in.GetId()), Subject: "Customer.GetCustomer"})
	if err != nil {
		panic(err)
	}

	out, ok := s.Customers[in.GetId()]
	if !ok {
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no customer with Id: %v", in.GetId()), Subject: "Customer.GetCustomer"})
		if err != nil {
			panic(err)
		}
		log.Printf("no customer with Id: %v", in.GetId())

		return &api.CustomerReply{}, status.Error(codes.NotFound, fmt.Sprintf("id: %v  was not found", in.GetId()))
	}

	log.Printf("successfully loaded customer of: id: %v, name: %v, address: %v", in.GetId(), out.GetName(), out.GetAddress())
	return &api.CustomerReply{Id: s.CustomerID, Name: out.GetName(), Address: out.GetAddress()}, nil
}
