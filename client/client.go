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
	"google.golang.org/grpc/status"
)

type Client struct {
	Nats  *nats.EncodedConn
	Redis *redis.Client
}

func (c *Client) Scenarios(scenario string) {
	log.Printf("Run scenario: %s ", scenario)
	err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("run scenario: %s", scenario), Subject: "Client.Scenarios"})
	if err != nil {
		panic(err)
	}

	switch scenario {
	case "s1":
		c.scenario1()
	case "s2":
		c.scenario2()

	default:
		log.Fatalf("no scenario called: %s", scenario)
		err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no scenario called: %s", scenario), Subject: "Client.Scenarios"})
		if err != nil {
			panic(err)
		}
	}
}

// 1. Bestellung lagernder Produkte durch einen Neukunden bis zum Verschicken.
// 2. Bestellung von drei Produkten, von denen nur eines auf Lager ist, bis zum Verschicken.
//      Die beiden nicht lagernden Produkte werden zu verschiedenen Zeit- punkten, durch verschiedene Zulieferer, geliefert.
// 3.  Stornieren einer Bestellung, die noch nicht verschickt wurde.
// 4. Empfangen eines defekten Artikels aus einer Bestellung mit mehreren Artikeln als Retoure mit sofortiger Ersatzlieferung.
// 5. Empfangen eines defekten Artikels aus einer Bestellung mit mehreren Artikeln als Retoure mit Rückbuchung des entsprechenden Teilbetrages.

func (c *Client) scenario1() {
	// Testen der einzelnen Komponenten
	// 1. Customer service
	// 2. Payment service
	// 3. Order service
	// 4. Catalog service

	////////////////////////////
	// Kommunikation mit Customer:
	////////////////////////////
	// Mithilfe von Redis Verbindung zu Customer aufbauen
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

	// - Neuen Kunden erstellen
	customer_r, customer_err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Simon", Address: "Munich"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	customer_r, customer_err = customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Max", Address: "Berlin"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	// - Kunde über ID anfordern
	customer_r, customer_err = customer.GetCustomer(customer_ctx, &api.GetCustomerRequest{Id: 3})
	if customer_err != nil {
		st, ok := status.FromError(customer_err)
		if !ok {
			log.Fatalf("An unexpected error occurred: %v", customer_err)
		}
		log.Printf("GetCustomer failed: %v", st.Message())
	}
	log.Printf("Got customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	////////////////////////////
	// Kommunikation mit Payment:
	////////////////////////////
	// Mithilfe von Redis Verbindung zu Payment aufbauen

	payment_redisVal := c.Redis.Get(context.TODO(), "payment")
	if payment_redisVal == nil {
		log.Fatal("service not registered")
	}
	payment_address, err := payment_redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	payment_conn, err := grpc.Dial(payment_address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer payment_conn.Close()
	payment := api.NewPaymentClient(payment_conn)
	payment_ctx, payment_cancel := context.WithTimeout(context.Background(), time.Second)
	defer payment_cancel()

	// - Neues Payment erstellen
	newPayment := &api.NewPaymentRequest{OrderId: 1, Value: 33.33}
	err = c.Nats.Publish("payment.new", newPayment)
	if err != nil {
		panic(err)
	}
	log.Printf("created payment: orderId:%v, value:%v", newPayment.GetOrderId(), newPayment.GetValue())

	// - Payment bezahlen
	payment_r, payment_err := payment.PayPayment(payment_ctx, &api.PayPaymentRequest{OrderId: newPayment.OrderId, Value: 33.98})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("payed payment: orderId:%v, value:%v", payment_r.GetOrderId(), payment_r.GetValue())

	// -  Payment zurückerstatten
	refundPayment := &api.RefundPaymentRequest{OrderId: 1, Value: 33.33}
	err = c.Nats.Publish("payment.refund", refundPayment)
	if err != nil {
		panic(err)
	}
	log.Printf("refund payment: orderId:%v, value:%v", refundPayment.GetOrderId(), refundPayment.GetValue())

	////////////////////////////
	// Kommunikation mit Order:
	////////////////////////////
	order_redisVal := c.Redis.Get(context.TODO(), "order")
	if order_redisVal == nil {
		log.Fatal("service not registered")
	}
	order_address, err := order_redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	order_conn, err := grpc.Dial(order_address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer order_conn.Close()

	order := api.NewOrderClient(order_conn)
	order_ctx, order_cancel := context.WithTimeout(context.Background(), time.Second)
	defer order_cancel()

	m := make(map[uint32]uint32)
	m[uint32(100)] = uint32(2)
	m[uint32(100)] = uint32(1)

	// Order erstellen (direkt)
	order_r, order_err := order.NewOrder(order_ctx, &api.NewOrderRequest{CustomerID: 1, Article: m})
	if order_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("created order: OrderId:%v", order_r.GetOrderId())

	// Order als bezahlt markieren (indirekt)
	paymentUpdate := &api.OrderPaymentUpdate{OrderId: 1}
	err = c.Nats.Publish("order.payment", paymentUpdate)
	if err != nil {
		panic(err)
	}
	log.Printf("updated order payment: orderId:%v", paymentUpdate.GetOrderId())

	time.Sleep(8 * time.Second)

	// Order als verschifft markieren (indirekt)
	shipmentUpdate := &api.OrderShipmentUpdate{OrderId: 1}
	err = c.Nats.Publish("order.shipment", shipmentUpdate)
	if err != nil {
		panic(err)
	}
	log.Printf("updated order shipment: orderId:%v", paymentUpdate.GetOrderId())

	time.Sleep(8 * time.Second)

	////////////////////////////
	// Kommunikation mit Payment:
	////////////////////////////
	// Mithilfe von Redis Verbindung zu Payment aufbauen

	catalog_redisVal := c.Redis.Get(context.TODO(), "catalog")
	if catalog_redisVal == nil {
		log.Fatal("service not registered")
	}
	catalog_address, err := catalog_redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	catalog_conn, err := grpc.Dial(catalog_address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer catalog_conn.Close()

	catalog := api.NewCatalogClient(catalog_conn)
	catalog_ctx, catalog_cancel := context.WithTimeout(context.Background(), time.Second)
	defer catalog_cancel()

	// - Neuen Eintrag erstellen
	catalog_r, catalog_err := catalog.NewCatalogArticle(catalog_ctx, &api.NewCatalog{Name: "Printer123", Description: "Very good printer!", Price: 102.50})
	if catalog_err != nil {
		log.Fatalf("Direct communication with catalog failed: %v", catalog_r)
	}
	log.Printf("Created catalog entry: Name:%v, Description:%v, Price:%v, Id:%v", catalog_r.GetName(), catalog_r.GetDescription(), catalog_r.GetPrice(), catalog_r.GetId())

	catalog_r, catalog_err = catalog.NewCatalogArticle(catalog_ctx, &api.NewCatalog{Name: "Lamp", Description: "Ligths up the room", Price: 42.00})
	if catalog_err != nil {
		log.Fatalf("Direct communication with catalog failed: %v", catalog_r)
	}
	log.Printf("Created catalog entry: Name:%v, Description:%v, Price:%v, Id:%v", catalog_r.GetName(), catalog_r.GetDescription(), catalog_r.GetPrice(), catalog_r.GetId())

	// - Kunde über ID anfordern
	catalog_r2, catalog_err := catalog.GetCatalogInfo(catalog_ctx, &api.GetCatalog{Id: 2})
	if catalog_err != nil {
		st, ok := status.FromError(catalog_err)
		if !ok {
			log.Fatalf("An unexpected error occurred: %v", catalog_err)
		}
		log.Printf("GetCatalogInfo failed: %v", st.Message())
	}
	log.Printf("Got catalog: Name:%v, Description:%v, Price:%v, Id:%v, Availability:%v", catalog_r2.GetName(), catalog_r2.GetDescription(), catalog_r2.GetPrice(), catalog_r2.GetId(), catalog_r2.GetAvailability())
}

func (c *Client) scenario2() {
	// Standartfall: Kund Bestellt lagernde Produkte
	// 1. Kunde erstellen
	// 2. Bestellung erstellen
	// 3. Auf Versandbestätigung warten

	log.Printf("Send xx to xx")

}
