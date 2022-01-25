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
	case "test":
		c.test()
	case "s1":
		c.scenario1()
	case "s2":
		c.scenario2()
	case "s3":
		c.scenario3()
	case "s4":
		c.scenario4()
	case "s5":
		c.scenario5()

	default:
		err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no scenario called: %s", scenario), Subject: "Client.Scenarios"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("no scenario called: %s", scenario)
	}
}

func (c *Client) getConnection(connectTo string) *grpc.ClientConn {
	redisVal := c.Redis.Get(context.TODO(), connectTo)
	if redisVal == nil {
		log.Fatalf("service %v not registered", connectTo)
	}
	address, err := redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return conn
}

func (c *Client) test() {
	// Testen der einzelnen Komponenten
	// 1. Customer service
	// 2. Payment service
	// 3. Catalog service
	// 4. Order service

	////////////////////////////
	// Kommunikation mit Customer:
	////////////////////////////

	// Mithilfe von Redis Verbindung zu Customer aufbauen
	customer_con := c.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

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

	// - Kunde über ID anfordern die nicht existiert
	_, customer_err = customer.GetCustomer(customer_ctx, &api.GetCustomerRequest{Id: 3})
	if customer_err != nil {
		st, ok := status.FromError(customer_err)
		if !ok {
			log.Fatalf("An unexpected error occurred: %v", customer_err)
		}
		log.Printf("GetCustomer failed: %v", st.Message())
	}

	customer_r, customer_err = customer.GetCustomer(customer_ctx, &api.GetCustomerRequest{Id: 1})
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
	payment_con := c.getConnection("payment")
	payment := api.NewPaymentClient(payment_con)
	payment_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer payment_con.Close()
	defer cancel()

	// - Neues Payment erstellen
	newPayment := &api.NewPaymentRequest{OrderId: 1, TotalCost: 33.33}
	err := c.Nats.Publish("payment.new", newPayment)
	if err != nil {
		panic(err)
	}
	log.Printf("created payment: orderId:%v, value:%v", newPayment.GetOrderId(), newPayment.GetTotalCost())

	// - Payment bezahlen (nicht komplett)
	payment_r, payment_err := payment.PayPayment(payment_ctx, &api.PayPaymentRequest{OrderId: newPayment.OrderId, Value: 10})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("payed payment: orderId:%v, stillToPay:%v", payment_r.GetOrderId(), payment_r.GetStillToPay())

	// - Payment stornieren
	cancelPayment := &api.CancelPaymentRequest{OrderId: 1, CustomerName: "Name1", CustomerAddress: "straße2"}
	err = c.Nats.Publish("payment.cancel", cancelPayment)
	if err != nil {
		panic(err)
	}
	log.Printf("cancel payment: orderId:%v", cancelPayment.GetOrderId())

	// -  Payment zurückerstatten
	refundPayment := &api.RefundPaymentRequest{OrderId: 1, CustomerName: customer_r.GetName(), CustomerAddress: customer_r.GetAddress(), Value: 33.33}
	err = c.Nats.Publish("payment.refund", refundPayment)
	if err != nil {
		panic(err)
	}
	log.Printf("refund payment: orderId:%v, value:%v", refundPayment.GetOrderId(), refundPayment.GetValue())

	////////////////////////////
	// Kommunikation mit Catalog:
	////////////////////////////

	// Mithilfe von Redis Verbindung zu Catalog aufbauen
	catalog_con := c.getConnection("catalog")
	catalog := api.NewCatalogClient(catalog_con)
	catalog_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer catalog_con.Close()
	defer cancel()

	// - Neuen Artikel hinzufügen
	catalog_r, catalog_err := catalog.NewCatalogArticle(catalog_ctx, &api.NewCatalog{Name: "Printer123", Description: "Very good printer!", Price: 102.50})
	if catalog_err != nil {
		log.Fatalf("Direct communication with catalog failed: %v", catalog_r)
	}
	log.Printf("Created catalog entry: Name:%v, Description:%v, Price:%v, Id:%v", catalog_r.GetName(), catalog_r.GetDescription(), catalog_r.GetPrice(), catalog_r.GetId())
	catalog_r, catalog_err = catalog.NewCatalogArticle(catalog_ctx, &api.NewCatalog{Name: "Lamp", Description: "Lights up the room", Price: 42.00})
	if catalog_err != nil {
		log.Fatalf("Direct communication with catalog failed: %v", catalog_r)
	}
	log.Printf("Created catalog entry: Name:%v, Description:%v, Price:%v, Id:%v", catalog_r.GetName(), catalog_r.GetDescription(), catalog_r.GetPrice(), catalog_r.GetId())

	// - Artikel Informationen anfordern
	catalog_r2, catalog_err := catalog.GetCatalogInfo(catalog_ctx, &api.GetCatalog{Id: 2})
	if catalog_err != nil {
		st, ok := status.FromError(catalog_err)
		if !ok {
			log.Fatalf("An unexpected error occurred: %v", catalog_err)
		}
		log.Printf("GetCatalogInfo failed: %v", st.Message())
	}
	log.Printf("Got catalog: Name:%v, Description:%v, Price:%v, Id:%v, Availability:%v", catalog_r2.GetName(), catalog_r2.GetDescription(), catalog_r2.GetPrice(), catalog_r2.GetId(), catalog_r2.GetAvailability())

	////////////////////////////
	// Kommunikation mit Order:
	////////////////////////////

	// Mithilfe von Redis Verbindung zu Order aufbauen
	order_con := c.getConnection("order")
	order := api.NewOrderClient(order_con)
	order_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer catalog_con.Close()
	defer cancel()

	m := make(map[uint32]uint32)
	m[uint32(1)] = uint32(2)
	m[uint32(2)] = uint32(1)

	// Order erstellen (direkt)
	order_r, order_err := order.NewOrder(order_ctx, &api.NewOrderRequest{CustomerID: 1, Articles: m})
	if order_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("created order: OrderId:%v", order_r.GetOrderId())

	time.Sleep(3 * time.Second)

	// Order als bezahlt markieren (indirekt)
	paymentUpdate := &api.OrderPaymentUpdate{OrderId: 1}
	err = c.Nats.Publish("order.payment", paymentUpdate)
	if err != nil {
		panic(err)
	}
	log.Printf("updated order payment: orderId:%v", paymentUpdate.GetOrderId())

	time.Sleep(3 * time.Second)

	// Order als verschifft markieren (indirekt)
	shipmentUpdate := &api.OrderShipmentUpdate{OrderId: 1}
	err = c.Nats.Publish("order.shipment", shipmentUpdate)
	if err != nil {
		panic(err)
	}
	log.Printf("updated order shipment: orderId:%v", paymentUpdate.GetOrderId())

	time.Sleep(3 * time.Second)

	// Artikel zurücksenden 1
	refundArticle := &api.RefundArticleRequest{OrderId: 1, ArticleId: 1}
	err = c.Nats.Publish("order.refund", refundArticle)
	if err != nil {
		panic(err)
	}
	log.Printf("refund article of: orderId:%v, articleID: %v", refundArticle.GetOrderId(), refundArticle.GetArticleId())

	// Artikel zurücksenden 2
	refundArticle = &api.RefundArticleRequest{OrderId: 1, ArticleId: 1}
	err = c.Nats.Publish("order.refund", refundArticle)
	if err != nil {
		panic(err)
	}
	log.Printf("refund article of: orderId:%v, articleID: %v", refundArticle.GetOrderId(), refundArticle.GetArticleId())

	time.Sleep(3 * time.Second)

	// Order stornieren 1
	cancelOrder := &api.CancelOrderRequest{OrderId: 1}
	err = c.Nats.Publish("order.cancel", cancelOrder)
	if err != nil {
		panic(err)
	}
	log.Printf("canceled order: orderId:%v", cancelOrder.GetOrderId())
}

func (c *Client) scenario1() {
	// 1. Bestellung lagernder Produkte durch einen Neukunden bis zum Verschicken.
	log.Printf("Scenario 1: Order of products in stock by a new customer ")
	err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("Scenario1: Order of products in stock by a new customer "), Subject: "Client.Scenario1"})
	if err != nil {
		panic(err)
	}

	////
	// Verbindung zu Catalog-Service aufbauen
	////
	catalog_con := c.getConnection("catalog")
	catalog := api.NewCatalogClient(catalog_con)
	catalog_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer catalog_con.Close()
	defer cancel()

	// - Neuen Artikel hinzufügen
	catalog_r1, catalog_err := catalog.NewCatalogArticle(catalog_ctx, &api.NewCatalog{Name: "Printer123", Description: "Very good printer!", Price: 102.50})
	if catalog_err != nil {
		log.Fatalf("Direct communication with catalog failed: %v", catalog_r1)
	}
	log.Printf("Created catalog entry: Name:%v, Description:%v, Price:%v, Id:%v", catalog_r1.GetName(), catalog_r1.GetDescription(), catalog_r1.GetPrice(), catalog_r1.GetId())
	catalog_r2, catalog_err := catalog.NewCatalogArticle(catalog_ctx, &api.NewCatalog{Name: "Lamp", Description: "Lights up the room", Price: 42.00})
	if catalog_err != nil {
		log.Fatalf("Direct communication with catalog failed: %v", catalog_r2)
	}
	log.Printf("Created catalog entry: Name:%v, Description:%v, Price:%v, Id:%v", catalog_r2.GetName(), catalog_r2.GetDescription(), catalog_r2.GetPrice(), catalog_r2.GetId())

	// - Bestand einbuchen
	addArticle := &api.AddStockRequest{Id: catalog_r1.Id, Amount: 3}
	err = c.Nats.Publish("stock.add", addArticle)
	if err != nil {
		panic(err)
	}
	log.Printf("Added amount %v to stock of article %v", 3, catalog_r1.GetId())
	addArticle = &api.AddStockRequest{Id: catalog_r2.Id, Amount: 5}
	err = c.Nats.Publish("stock.add", addArticle)
	if err != nil {
		panic(err)
	}
	log.Printf("Added amount %v to stock of article %v", 5, catalog_r2.GetId())

	////
	// Verbindung zu Customer-Service aufbauen
	////
	customer_con := c.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	// Neuen Customer anlegen
	customer_r, customer_err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Max", Address: "Berlin"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	////
	// Verbindung zu Order-Service aufbauen
	////
	order_con := c.getConnection("order")
	order := api.NewOrderClient(order_con)
	order_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer catalog_con.Close()
	defer cancel()

	// Neue Order anlegen
	articles := make(map[uint32]uint32)
	articles[catalog_r1.GetId()] = uint32(1)
	articles[catalog_r2.GetId()] = uint32(1)

	order_r, order_err := order.NewOrder(order_ctx, &api.NewOrderRequest{CustomerID: customer_r.GetId(), Articles: articles})
	if order_err != nil {
		log.Fatalf("Direct communication with order failed: %v", order_r)
	}
	log.Printf("Created order: OrderId:%v OrderCost: %v", order_r.GetOrderId(), order_r.GetTotalCost())

	////
	// Verbindung zu Payment-Service aufbauen
	////
	payment_con := c.getConnection("payment")
	payment := api.NewPaymentClient(payment_con)
	payment_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer payment_con.Close()
	defer cancel()

	// Payment bezahlen
	payment_r, payment_err := payment.PayPayment(payment_ctx, &api.PayPaymentRequest{OrderId: order_r.GetOrderId(), Value: order_r.GetTotalCost()})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("Payed payment: orderId:%v, value:%v", payment_r.GetOrderId(), order_r.GetTotalCost())
}

func (c *Client) scenario2() {

	log.Printf("Scenario 2: Ordering three products, only one of which is in stock")
	err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("Scenario2: Ordering three products, only one of which is in stock"), Subject: "Client.Scenario2"})
	if err != nil {
		panic(err)
	}

	// Stock und Catalog "befüllen"
	log.Printf("Created catalog and fill up stock")
	err = c.Nats.Publish("catalog.first", "")
	if err != nil {
		panic(err)
	}

	////
	// Verbindung zu Customer-Service aufbauen
	////
	customer_con := c.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	customer_r, customer_err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Simon", Address: "Munich"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	////
	// Verbindung zu Order-Service aufbauen
	////
	order_con := c.getConnection("order")
	order := api.NewOrderClient(order_con)
	order_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer order_con.Close()
	defer cancel()

	// Neue Order anlegen
	// Bestellung von Artikel 1, 3, 4
	articles := make(map[uint32]uint32)
	articles[1] = uint32(1)
	articles[3] = uint32(1)
	articles[4] = uint32(1)

	log.Printf("articles:%v", articles)

	order_r, order_err := order.NewOrder(order_ctx, &api.NewOrderRequest{CustomerID: customer_r.GetId(), Articles: articles})
	if order_err != nil {
		log.Fatalf("Direct communication with order failed: %v", order_r)
	}
	log.Printf("Created order: OrderId:%v OrderCost: %v", order_r.GetOrderId(), order_r.GetTotalCost())

	////
	// Verbindung zu Payment-Service aufbauen
	////
	payment_con := c.getConnection("payment")
	payment := api.NewPaymentClient(payment_con)
	payment_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer payment_con.Close()
	defer cancel()

	// Payment bezahlen
	payment_r, payment_err := payment.PayPayment(payment_ctx, &api.PayPaymentRequest{OrderId: order_r.GetOrderId(), Value: order_r.GetTotalCost()})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("Payed payment: orderId:%v, value:%v", payment_r.GetOrderId(), order_r.GetTotalCost())

	////
	// Verbindung zu Supplier-Service aufbauen
	////
	supplier_con := c.getConnection("supplier")
	supplier := api.NewSupplierClient(supplier_con)
	supplier_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer supplier_con.Close()
	defer cancel()

	// Lieferung des Artikels der nicht mehr vorrätig war
	supplier_r, supplier_err := supplier.DeliveredArticle(supplier_ctx, &api.NewArticles{OrderId: order_r.GetOrderId(), ArticleId: 3, Amount: 1, NameSupplier: "Supplier1"})
	if supplier_err != nil {
		log.Fatalf("Direct communication with supplier failed: %v", supplier_r)
	}
	log.Printf("Delivered articles: orderId:%v, articleId:%v, amount:%v, name supplier:%v", supplier_r.GetOrderId(), supplier_r.GetArticleId(), supplier_r.GetAmount(), supplier_r.GetNameSupplier())

	// Lieferung des Artikels der nicht mehr vorrätig war
	supplier_r, supplier_err = supplier.DeliveredArticle(supplier_ctx, &api.NewArticles{OrderId: order_r.GetOrderId(), ArticleId: 4, Amount: 1, NameSupplier: "Supplier2"})
	if supplier_err != nil {
		log.Fatalf("Direct communication with supplier failed: %v", supplier_r)
	}
	log.Printf("Delivered articles: orderId:%v, articleId:%v, amount:%v, name supplier:%v", supplier_r.GetOrderId(), supplier_r.GetArticleId(), supplier_r.GetAmount(), supplier_r.GetNameSupplier())

}
func (c *Client) scenario3() {
	log.Printf("Scenario 3: Cancel an order that has not been shipped yet")
	err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("Scenario 3: Cancel an order that has not been shipped yet"), Subject: "Client.Scenario3"})
	if err != nil {
		panic(err)
	}

	// Stock und Catalog "befüllen"
	log.Printf("Created catalog and fill up stock")
	err = c.Nats.Publish("catalog.first", "")
	if err != nil {
		panic(err)
	}

	////
	// Verbindung zu Customer-Service aufbauen
	////
	customer_con := c.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	customer_r, customer_err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Peter", Address: "Ulm"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	////
	// Verbindung zu Order-Service aufbauen
	////
	order_con := c.getConnection("order")
	order := api.NewOrderClient(order_con)
	order_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer order_con.Close()
	defer cancel()

	// Neue Order anlegen
	// Bestellung von mehrere nicht lagernden Artikel
	articles := make(map[uint32]uint32)
	articles[1] = uint32(2)
	articles[2] = uint32(42)
	articles[4] = uint32(33)

	order_r, order_err := order.NewOrder(order_ctx, &api.NewOrderRequest{CustomerID: customer_r.GetId(), Articles: articles})
	if order_err != nil {
		log.Fatalf("Direct communication with order failed: %v", order_r)
	}
	log.Printf("Created order: OrderId:%v OrderCost: %v", order_r.GetOrderId(), order_r.GetTotalCost())

	////
	// Verbindung zu Payment-Service aufbauen
	////
	payment_con := c.getConnection("payment")
	payment := api.NewPaymentClient(payment_con)
	payment_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer payment_con.Close()
	defer cancel()

	// Payment bezahlen
	payment_r, payment_err := payment.PayPayment(payment_ctx, &api.PayPaymentRequest{OrderId: order_r.GetOrderId(), Value: order_r.GetTotalCost()})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("Payed payment: orderId:%v, value:%v", payment_r.GetOrderId(), order_r.GetTotalCost())

	time.Sleep(5 * time.Second)

	// Order stornieren
	cancelOrder := &api.CancelOrderRequest{OrderId: 1}
	err = c.Nats.Publish("order.cancel", cancelOrder)
	if err != nil {
		panic(err)
	}
	log.Printf("Canceled order: orderId:%v", cancelOrder.GetOrderId())
}

func (c *Client) scenario4() {
	log.Printf("Scenario 4: Return of a defective article and sending a replacement")
	err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("Scenario4: Return of a defective article and sending a replacement"), Subject: "Client.Scenario4"})
	if err != nil {
		panic(err)
	}

	// Stock und Catalog "befüllen"
	err = c.Nats.Publish("catalog.first", "")
	if err != nil {
		panic(err)
	}

	////
	// Verbindung zu Customer-Service aufbauen
	////
	customer_con := c.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	customer_r, customer_err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Simon", Address: "Munich"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	////
	// Verbindung zu Order-Service aufbauen
	////
	order_con := c.getConnection("order")
	order := api.NewOrderClient(order_con)
	order_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer order_con.Close()
	defer cancel()

	// Neue Order anlegen
	articles := make(map[uint32]uint32)
	articles[1] = uint32(2)
	articles[2] = uint32(1)

	order_r, order_err := order.NewOrder(order_ctx, &api.NewOrderRequest{CustomerID: customer_r.GetId(), Articles: articles})
	if order_err != nil {
		log.Fatalf("Direct communication with order failed: %v", order_r)
	}
	log.Printf("Created order: OrderId:%v OrderCost: %v", order_r.GetOrderId(), order_r.GetTotalCost())

	////
	// Verbindung zu Payment-Service aufbauen
	////
	payment_con := c.getConnection("payment")
	payment := api.NewPaymentClient(payment_con)
	payment_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer payment_con.Close()
	defer cancel()

	// Payment bezahlen
	payment_r, payment_err := payment.PayPayment(payment_ctx, &api.PayPaymentRequest{OrderId: order_r.GetOrderId(), Value: order_r.GetTotalCost()})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("Payed payment: orderId:%v, value:%v", payment_r.GetOrderId(), order_r.GetTotalCost())

	// Warten, damit Ware versendet wird
	time.Sleep(3 * time.Second)

	////
	// Verbindung zu Shipment-Service aufbauen
	////
	shipment_con := c.getConnection("shipment")
	shipment := api.NewShipmentClient(shipment_con)
	shipment_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer shipment_con.Close()
	defer cancel()

	// Rückgabe eines defekten Artikels mit Ersatz
	shipment_r, shipment_err := shipment.ReturnDefectArticle(shipment_ctx, &api.ShipmentReturnRequest{Id: order_r.GetOrderId(), ArticleId: 1, Amount: 1})
	if shipment_err != nil {
		log.Fatalf("Direct communication with shipment failed: %v", shipment_r)
	}
	log.Printf("Returned articles: orderId:%v, articleId:%v, amount:%v", shipment_r.GetId(), shipment_r.GetArticleId(), shipment_r.GetAmount())

}
func (c *Client) scenario5() {
	log.Printf("Scenario 5: Return of a defective article and refunding it")
	err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("Scenario5: Return of a defective article and refunding it"), Subject: "Client.Scenario5"})
	if err != nil {
		panic(err)
	}

	// Stock und Catalog "befüllen"
	err = c.Nats.Publish("catalog.first", "")
	if err != nil {
		panic(err)
	}

	////
	// Verbindung zu Customer-Service aufbauen
	////
	customer_con := c.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	customer_r, customer_err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Simon", Address: "Munich"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	////
	// Verbindung zu Order-Service aufbauen
	////
	order_con := c.getConnection("order")
	order := api.NewOrderClient(order_con)
	order_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer order_con.Close()
	defer cancel()

	// Neue Order anlegen
	articles := make(map[uint32]uint32)
	articles[1] = uint32(2)
	articles[2] = uint32(1)

	order_r, order_err := order.NewOrder(order_ctx, &api.NewOrderRequest{CustomerID: customer_r.GetId(), Articles: articles})
	if order_err != nil {
		log.Fatalf("Direct communication with order failed: %v", order_r)
	}
	log.Printf("Created order: OrderId:%v OrderCost: %v", order_r.GetOrderId(), order_r.GetTotalCost())

	////
	// Verbindung zu Payment-Service aufbauen
	////
	payment_con := c.getConnection("payment")
	payment := api.NewPaymentClient(payment_con)
	payment_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer payment_con.Close()
	defer cancel()

	// Payment bezahlen
	payment_r, payment_err := payment.PayPayment(payment_ctx, &api.PayPaymentRequest{OrderId: order_r.GetOrderId(), Value: order_r.GetTotalCost()})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("Payed payment: orderId:%v, value:%v", payment_r.GetOrderId(), order_r.GetTotalCost())

	// Warten, damit Ware versendet wird
	time.Sleep(3 * time.Second)

	////
	// Verbindung zu Shipment-Service aufbauen
	////
	shipment_con := c.getConnection("shipment")
	shipment := api.NewShipmentClient(shipment_con)
	shipment_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer shipment_con.Close()
	defer cancel()

	// Rückgabe eines defekten Artikels mit Rückbuchung
	shipment_r, shipment_err := shipment.Refund(shipment_ctx, &api.ShipmentReturnRequest{Id: order_r.GetOrderId(), ArticleId: 1, Amount: 1})
	if shipment_err != nil {
		log.Fatalf("Direct communication with shipment failed: %v", shipment_r)
	}
	log.Printf("Returned articles: orderId:%v, articleId:%v, amount:%v", shipment_r.GetId(), shipment_r.GetArticleId(), shipment_r.GetAmount())
}
