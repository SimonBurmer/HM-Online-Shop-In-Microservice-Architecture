package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/shipment"
	"google.golang.org/grpc"
)

const (
	port = ":50056"
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
			err = rdb.Set(context.TODO(), "shipment", "shipment-service"+port, 13*time.Second).Err()
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
	shipmentServer := shipment.Server{Nats: c, Shipment: make(map[uint32]*api.NewShipmentRequest), ShipmentID: 0}
	api.RegisterShipmentServer(s, &shipmentServer)

	// Subscribed to NATS Channels
	newShipmentSubscription, err := c.Subscribe("shipment.new", func(msg *api.NewShipmentRequest) {
		shipmentServer.NewShipment(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer newShipmentSubscription.Unsubscribe()

	gotStockSubscription, err := c.Subscribe("shipment.articles", func(msg *api.ShipmentReadiness) {
		shipmentServer.ShipmentReady(msg)
	})
	if err != nil {
		log.Fatal("cannot subscribe")
	}
	defer gotStockSubscription.Unsubscribe()

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
