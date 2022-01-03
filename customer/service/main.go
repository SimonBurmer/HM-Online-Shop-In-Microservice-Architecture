package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/customer"
	"google.golang.org/grpc"
)

const (
	port = ":50055"
)

func main() {
	// Definiere Port der "abhört" wird
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// Verbindung zu Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "host.docker.internal:6379",
		Password: "", // no password set
	})

	// Registration im Redis
	go func() {
		for {
			err = rdb.Set(context.TODO(), "customer", "host.docker.internal"+port, 13*time.Second).Err()
			if err != nil {
				panic(err)
			}
			log.Print("register service")
			time.Sleep(10 * time.Second)
		}
	}()

	// Verbindung zu NATS
	nc, err := nats.Connect("host.docker.internal:4222")
	if err != nil {
		log.Fatal(err)
	}
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Nats Channel subscriben
	subscription, err := c.Subscribe("customer", func(p *api.NewCustomerRequest) {
		log.Printf("Received newCustomerRequest with: name: %v ,adress: %v", p.GetName(),p.GetAddress())
		
	})
	if err != nil {
		log.Fatalf("cannot subscribe: %v", err)
	}
	defer subscription.Unsubscribe()

	// Erzeugt den fertigen Service
	api.RegisterCustomerServer(s, &customer.Server{Nats: c, Customers: make(map[uint32]*api.NewCustomerRequest), CustomerID: 0})
	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
