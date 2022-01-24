package order

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	Nats    *nats.EncodedConn
	Redis   *redis.Client
	Orders  map[uint32]*api.OrderStorage
	OrderID uint32
	api.UnimplementedOrderServer
}

func (s *Server) NewOrder(ctx context.Context, in *api.NewOrderRequest) (*api.OrderReply, error) {
	log.Printf("received new order request of: customerID: %v", in.GetCustomerID())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received new order request of: customerID: %v", in.GetCustomerID()), Subject: "Order.NewOrder"})
	if err != nil {
		panic(err)
	}

	// Verbindung zu Customer-Service aufbauen
	customer_con := s.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	// Überprüfen ob Customer existiert
	_, customer_err := customer.GetCustomer(customer_ctx, &api.GetCustomerRequest{Id: in.CustomerID})
	if customer_err != nil {
		st, ok := status.FromError(customer_err)
		if !ok {
			log.Fatalf("An unexpected error occurred: %v", customer_err)
		}
		log.Printf("GetCustomer failed: %v", st.Message())
		return &api.OrderReply{}, status.Error(codes.NotFound, "Given Customer was not found")
	}

	// Verbindung zu Catalog-Service aufbauen
	catalog_con := s.getConnection("catalog")
	catalog := api.NewCatalogClient(catalog_con)
	catalog_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer catalog_con.Close()
	defer cancel()

	// Überprüfen ob bestellte Artikel existieren und Bestellsumme berechnen
	var totalCost float64
	for articleID, articleAmount := range in.GetArticles() {
		article, catalog_err := catalog.GetCatalogInfo(catalog_ctx, &api.GetCatalog{Id: articleID})
		if catalog_err != nil {
			st, ok := status.FromError(customer_err)
			if !ok {
				log.Fatalf("An unexpected error occurred: %v", customer_err)
			}
			log.Printf("GetCatalogInfo failed: %v", st.Message())
			return &api.OrderReply{}, status.Error(codes.NotFound, fmt.Sprintf("Given article was not found: articleID %v", articleID))
		}
		totalCost = totalCost + article.GetPrice()*float64(articleAmount)
	}

	// Neue Order abspeichern
	s.OrderID = s.OrderID + 1
	s.Orders[s.OrderID] = &api.OrderStorage{CustomerID: in.CustomerID, Articles: in.Articles, TotalCost: float32(totalCost), Payed: false, Shipped: false, Canceled: false}

	// Neues Payment erstellen
	newPayment := &api.NewPaymentRequest{OrderId: s.OrderID, Value: float32(totalCost)}
	err = s.Nats.Publish("payment.new", newPayment)
	if err != nil {
		panic(err)
	}

	log.Printf("successfully created new order with orderId: %v", s.OrderID)
	return &api.OrderReply{OrderId: s.OrderID, TotalCost: float32(totalCost)}, nil
}

func (s *Server) OrderPaymentUpdate(in *api.OrderPaymentUpdate) {
	log.Printf("received order payment update of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received order payment update of: orderId: %v", in.GetOrderId()), Subject: "Order.OrderPaymentUpdate"})
	if err != nil {
		panic(err)
	}

	// Order laden und auf bezahlt setzen
	out := s.getOrder(in.GetOrderId())
	out.Payed = true
	s.Orders[in.GetOrderId()] = out

	// - Verbindung zu Customer-Service aufbauen
	customer_con := s.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	// - Kunden-Informationen holen
	customer_r, customer_err := customer.GetCustomer(customer_ctx, &api.GetCustomerRequest{Id: out.GetCustomerID()})
	if customer_err != nil {
		log.Fatalf("could not get customer of: customerId: %v", out.GetCustomerID())
	}

	// Neues Shipment erstellen
	newShipment := &api.NewShipmentRequest{OrderID: in.GetOrderId(), Articles: out.GetArticles(), Address: customer_r.GetAddress()}
	err = s.Nats.Publish("shipment.new", newShipment)
	if err != nil {
		panic(err)
	}

	log.Printf("order with orderId %v has been payed!", in.GetOrderId())
}

func (s *Server) OrderShipmentUpdate(in *api.OrderShipmentUpdate) {
	log.Printf("received order shipment update of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received order shipment update of: orderId: %v", in.GetOrderId()), Subject: "Order.OrderShipmentUpdate"})
	if err != nil {
		panic(err)
	}

	// Order laden und auf verschickt setzen
	out := s.getOrder(in.GetOrderId())
	out.Shipped = true
	s.Orders[in.GetOrderId()] = out
	log.Printf("order with orderId %v has been shipped!", in.GetOrderId())
}

func (s *Server) CancelOrderRequest(in *api.CancelOrderRequest) {
	log.Printf("received cancel order request of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received cancel order request of: orderId: %v", in.GetOrderId()), Subject: "Order.CancelOrderRequest"})
	if err != nil {
		panic(err)
	}

	// Order laden (checken ob Order mit gegebener Id existiert)
	out := s.getOrder(in.GetOrderId())

	// - Verbindung zu Customer-Service aufbauen
	customer_con := s.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	// - Kunden-Informationen holen
	customer_r, customer_err := customer.GetCustomer(customer_ctx, &api.GetCustomerRequest{Id: out.GetCustomerID()})
	if customer_err != nil {
		log.Fatalf("could not get customer of: customerId: %v", out.GetCustomerID())
	}

	// Payment der Order stornieren
	cancelPayment := &api.CancelPaymentRequest{OrderId: in.GetOrderId(), CustomerName: customer_r.GetName(), CustomerAddress: customer_r.GetAddress()}
	err = s.Nats.Publish("payment.cancel", cancelPayment)
	if err != nil {
		panic(err)
	}

	// Shipment der Order stornieren
	cancelShipment := &api.CancelShipmentRequest{Id: in.GetOrderId()}
	err = s.Nats.Publish("shipment.cancel", cancelShipment)
	if err != nil {
		panic(err)
	}

	// Stornierte Order zurückzahlen
	refundPayment := &api.RefundPaymentRequest{OrderId: in.GetOrderId(), CustomerName: customer_r.GetName(), CustomerAddress: customer_r.GetAddress(), Value: out.GetTotalCost()}
	err = s.Nats.Publish("payment.refund", refundPayment)
	if err != nil {
		panic(err)
	}

	// Order als storniert markieren
	out.Canceled = true
	s.Orders[in.GetOrderId()] = out

	log.Printf("successfully canceled order with orderId: %v", in.GetOrderId())
}

func (s *Server) RefundArticleRequest(in *api.RefundArticleRequest) {
	log.Printf("received refund article request of: orderId: %v, articleId: %v", in.GetOrderId(), in.GetArticleId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received refund article request of: orderId: %v, articleId: %v", in.GetOrderId(), in.GetArticleId()), Subject: "Order.RefundArticleRequest"})
	if err != nil {
		panic(err)
	}

	// Order laden (checken ob Order mit gegebener Id existiert)
	out := s.getOrder(in.GetOrderId())

	// Rücksendung aus Bestellung löschen
	delete(out.GetArticles(), in.GetArticleId())

	// Preis der Rücksendung erstatten
	// - Verbindung zu Catalog-Service aufbauen
	catalog_con := s.getConnection("catalog")
	catalog := api.NewCatalogClient(catalog_con)
	catalog_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer catalog_con.Close()
	defer cancel()

	// - Artikel-Informationen holen
	article, catalog_err := catalog.GetCatalogInfo(catalog_ctx, &api.GetCatalog{Id: in.GetArticleId()})
	if catalog_err != nil {
		log.Fatalf("could not get catalog info of: articleId: %v", in.GetArticleId())
	}

	// - Preis der Bestellung neu berechnen und abspeichern
	log.Printf("%v", s.getOrder(in.GetOrderId()).GetTotalCost())
	out.TotalCost = out.TotalCost - float32(article.GetPrice())
	s.Orders[in.GetOrderId()] = out
	log.Printf("%v", s.getOrder(in.GetOrderId()).GetTotalCost())

	// - Verbindung zu Customer-Service aufbauen
	customer_con := s.getConnection("customer")
	customer := api.NewCustomerClient(customer_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_con.Close()
	defer cancel()

	// - Kunden-Informationen holen
	customer_r, customer_err := customer.GetCustomer(customer_ctx, &api.GetCustomerRequest{Id: out.GetCustomerID()})
	if customer_err != nil {
		log.Fatalf("could not get customer of: customerId: %v", out.GetCustomerID())
	}

	// - Rückerstattung erstellen
	refundPayment := &api.RefundPaymentRequest{OrderId: in.GetOrderId(), CustomerName: customer_r.GetName(), CustomerAddress: customer_r.GetAddress(), Value: float32(article.GetPrice())}
	err = s.Nats.Publish("payment.refund", refundPayment)
	if err != nil {
		panic(err)
	}

	log.Printf("successfully refund articleID: %v of orderId: %v", in.GetArticleId(), in.GetOrderId())
}

func (s *Server) getConnection(connectTo string) *grpc.ClientConn {
	redisVal := s.Redis.Get(context.TODO(), connectTo)
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

func (s *Server) getOrder(orderId uint32) *api.OrderStorage {
	out, ok := s.Orders[orderId]
	if !ok {
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no order with orderId: %v", orderId), Subject: "Order.getOrder"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("no order with orderId: %v", orderId)
	}

	return out
}
