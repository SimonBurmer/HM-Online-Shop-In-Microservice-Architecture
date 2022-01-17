package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/stock"
	"google.golang.org/grpc"
)

const (
	port = ":50057"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// Verbindung zu Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
	})

	// Registration im Redis
	go func() {
		for {
			err = rdb.Set(context.TODO(), "stock", "host.docker.internal"+port, 13*time.Second).Err()
			if err != nil {
				panic(err)
			}
			log.Print("register service")
			time.Sleep(10 * time.Second)
		}
	}()

	// Verbindung zu NATS
	nc, err := nats.Connect("nats:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}

	// Erzeugt den fertigen service
	stockServer := stock.Server{Nats: nc, Redis: rdb, Stock: make(map[uint32]*api.NewStockRequest), StockID: 0}
	api.RegisterStockServer(s, &stockServer)
	//api.RegisterStockServer(s, &stock.Server{Nats: nc, Redis: rdb, Stock: make(map[uint32]*api.NewStockRequest), StockID: 0})

	// Subscribt einen Nats Channel
	newStockSubscription, err := c.Subscribe("stock.add", func(msg *api.AddStockRequest) {
		stockServer.AddStock(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer newStockSubscription.Unsubscribe()

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
