package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"google.golang.org/grpc"
)

type Client struct {
	Nats  *nats.Conn
	Redis *redis.Client
}

func (c *Client) Scenarios(scenario string) {
	log.Printf("Run scenario: %s ", scenario)
	err := c.Nats.Publish("log.client", []byte(fmt.Sprintf("Run scenario: %s", scenario)))
	if err != nil {
		panic(err)
	}

	switch scenario {
	case "s1":
		c.scenario1()
	case "s2":
		c.scenario2()

	default:
		log.Fatalf("No scenario called: %s", scenario)
		err := c.Nats.Publish("log.client", []byte(fmt.Sprintf("No scenario called: %s", scenario)))
		if err != nil {
			panic(err)
		}
	}
}

func (c *Client) scenario1() {
	// Standartfall: Kund Bestellt lagernde Produkte
	// 1. Kunde erstellen
	// 2. Bestellung erstellen
	// 3. Aufversandbestägung warten

	// Mithilfe Redis Customer conn aufbauen
	customer_redisVal := c.Redis.Get(context.TODO(), "customer")
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
	customer := api.NewCustomerClient(customer_conn)
	customer_ctx, customer_cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_cancel()

	// Direkte Kommunikation mit Customer:
	// - Neuen Kunden erstellen
	r, err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Simon", Address: "Munich"})
	if err != nil {
		log.Fatalf("Direct communication with customer failed: %v", r)
	}

	log.Printf("Created customer: %s,%s,%d", r.GetName(), r.GetAddress(), r.GetId())
	err = c.Nats.Publish("log.client", []byte(fmt.Sprintf("Created customer: %s,%s,%d", r.GetName(), r.GetAddress(), r.GetId())))
	if err != nil {
		panic(err)
	}

}

func (c *Client) scenario2() {
	log.Printf("Send xx to xx")

}
