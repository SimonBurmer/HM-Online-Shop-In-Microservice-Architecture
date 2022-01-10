package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	nats *nats.Conn
	api.UnimplementedGreeterServer
}

func (s server) SayHello(ctx context.Context, in *api.HelloRequest) (*api.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	// Indirekte Kommunikation über NATS
	err := s.nats.Publish("log.greeter", []byte(fmt.Sprintf("received: %v", in.GetName()))) // Stellt Logger einer Nachricht rein!!
	if err != nil {
		panic(err)
	}
	return &api.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// Registration im Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
	})

	go func() {
		for {
			err = rdb.Set(context.TODO(), "greeter", "127.0.0.1"+port, 13*time.Second).Err()
			if err != nil {
				panic(err)
			}
			log.Print("register service")
			time.Sleep(10 * time.Second)
		}
	}()

	// Verbindung zu NATS aufbauen
	nc, err := nats.Connect("nats:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Erzeugt den fertigen Service
	api.RegisterGreeterServer(s, &server{nats: nc})
	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
