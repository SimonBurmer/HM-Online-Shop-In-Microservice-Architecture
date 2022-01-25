package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/payment"
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
		Password: "", 
	})

	// Registration im Redis
	go func() {
		for {
			err = rdb.Set(context.TODO(), "payment", "payment-service"+port, 13*time.Second).Err()
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
	paymentServer := payment.Server{Nats: c, Payments: make(map[uint32]*api.PaymentStorage)}
	api.RegisterPaymentServer(s, &paymentServer)

	// Subscriben von Nats Channel
	newPaymentSubscription, err := c.Subscribe("payment.new", func(msg *api.NewPaymentRequest) {
		paymentServer.NewPayment(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer newPaymentSubscription.Unsubscribe()

	cancelPaymentSubscription, err := c.Subscribe("payment.cancel", func(msg *api.CancelPaymentRequest) {
		paymentServer.CancelPayment(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer cancelPaymentSubscription.Unsubscribe()

	refundPaymentSubscription, err := c.Subscribe("payment.refund", func(msg *api.RefundPaymentRequest) {
		paymentServer.RefundPayment(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer refundPaymentSubscription.Unsubscribe()

	// Startet Server
	err = s.Serve(lis)

	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
