package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/order"
	"google.golang.org/grpc"
)

const (
	port = ":50056"
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
		Addr:     "redis:6379",
		Password: "", // no password set
	})

	// Registration im Redis
	go func() {
		for {
			err = rdb.Set(context.TODO(), "order", "order-service"+port, 13*time.Second).Err()
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

	// Erzeugt den fertigen Service
	orderServer := order.Server{Nats: c, Redis: rdb, Orders: make(map[uint32]*api.OrderStorage), OrderID: 0}
	api.RegisterOrderServer(s, &orderServer)

	// Subscribt einen Nats Channel
	paymentUpdateSubscription, err := c.Subscribe("order.payment", func(msg *api.OrderPaymentUpdate) {
		orderServer.OrderPaymentUpdate(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer paymentUpdateSubscription.Unsubscribe()

	shipmentUpdateSubscription, err := c.Subscribe("order.shipment", func(msg *api.OrderShipmentUpdate) {
		orderServer.OrderShipmentUpdate(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer shipmentUpdateSubscription.Unsubscribe()

	cancleOrderSubscription, err := c.Subscribe("order.cancel", func(msg *api.CancelOrderRequest) {
		orderServer.CancelOrderRequest(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer cancleOrderSubscription.Unsubscribe()

	refundArticleSubscription, err := c.Subscribe("order.refund", func(msg *api.RefundArticleRequest) {
		orderServer.RefundArticleRequest(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer refundArticleSubscription.Unsubscribe()

	// Startet Server
	err = s.Serve(lis)

	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
