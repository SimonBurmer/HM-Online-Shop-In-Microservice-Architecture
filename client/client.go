package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"google.golang.org/grpc"
)

const (
	name = "World"
)

func main() {
	// Verbinden zu Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
	})

	// Von Redis Verbindung zu Greeter aufbauen
	redisVal := rdb.Get(context.TODO(), "greeter")
	if redisVal == nil {
		log.Fatal("service not registered")
	}
	address, err := redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Verbindung zu Greeter aufgebaut
	c := api.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Direkte Kommuniaktion mit Greeter
	r, err := c.SayHello(ctx, &api.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	///////////// CUSTOMER ///////////////
	print("ASDKJFKLSJDAFDSJAKFAJKSDFDSLA")

	// Von Redis Verbindung zu Greeter aufbauen
	customer_redisVal := rdb.Get(context.TODO(), "customer")
	if customer_redisVal == nil {
		log.Fatal("service not registered")
	}
	customer_address, err := customer_redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	customer_conn, err := grpc.Dial(customer_address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer customer_conn.Close()

	// Verbindung mit Customer
	customer := api.NewCustomerClient(customer_conn)

	customer_ctx, customer_cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_cancel()

	// Direkte Kommuniaktion mit Customer
	customer_r, err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "test", Address: "test"})
	if err != nil {
		log.Fatalf("direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
