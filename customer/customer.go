package customer

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats *nats.Conn
	api.UnimplementedGreeterServer
}

func (s Server) SayHello(ctx context.Context, in *api.HelloRequest) (*api.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	// Indirekte Kommunikation über NATS
	err := s.Nats.Publish("log.greeter", []byte(fmt.Sprintf("received: %v", in.GetName()))) // Stellt Logger einer Nachricht rein!!
	if err != nil {
		panic(err)
	}
	return &api.HelloReply{Message: "Hello " + in.GetName()}, nil
}
