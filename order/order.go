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

	// Verbindung zu Customer herstellen
	custoner_con := s.GetConnection("customer")
	customer := api.NewCustomerClient(custoner_con)
	customer_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer custoner_con.Close()
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

	// Prüfen ob bestelle artikel in catalog

	// gesamtsumme der bestellung von XXX
	totalCost := 3

	s.OrderID = s.OrderID + 1
	s.Orders[s.OrderID] = &api.OrderStorage{CustomerID: in.CustomerID, Article: in.Article, TotalCost: float32(totalCost), Payed: false, Shipped: false}
	log.Printf("successfully created new order with orderId: %v", s.OrderID)

	// Paymentservice reinstellen

	return &api.OrderReply{OrderId: s.OrderID}, nil
}

func (s *Server) OrderPaymentUpdate(in *api.OrderPaymentUpdate) {
	log.Printf("received order payment update of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received order payment update of: orderId: %v", in.GetOrderId()), Subject: "Order.OrderPaymentUpdate"})
	if err != nil {
		panic(err)
	}

	out, ok := s.Orders[in.GetOrderId()]
	if !ok {
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no order with orderId: %v", in.GetOrderId()), Subject: "Order.OrderPaymentUpdate"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("no order with orderId: %v", in.GetOrderId())
	}

	// Checken ob schon bezaht wurde

	out.Payed = true
	s.Orders[in.GetOrderId()] = out

	// Shipment reinstellen

	log.Printf("order with orderId %v has been payed!", in.GetOrderId())
}

func (s *Server) OrderShipmentUpdate(in *api.OrderShipmentUpdate) {
	log.Printf("received order shipment update of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received order shipment update of: orderId: %v", in.GetOrderId()), Subject: "Order.OrderShipmentUpdate"})
	if err != nil {
		panic(err)
	}

	out, ok := s.Orders[in.GetOrderId()]
	if !ok {
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no order with orderId: %v", in.GetOrderId()), Subject: "Order.OrderShipmentUpdate"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("no order with orderId: %v", in.GetOrderId())
	}

	// Checken ob schon verschifft wurde

	if !out.Payed {
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("Order shipped but not payed: orderId: %v", in.GetOrderId()), Subject: "Order.OrderShipmentUpdate"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("Order shipped but not payed: orderId: %v", in.GetOrderId())
	}

	out.Shipped = true
	s.Orders[in.GetOrderId()] = out

	log.Printf("order with orderId %v has been shipped!", in.GetOrderId())
}

func (s *Server) RefundOrderRequest(in *api.RefundOrderRequest) {
	log.Printf("order with orderId ")

	// TODO!!!

}

func (s *Server) GetConnection(connectTo string) *grpc.ClientConn {
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
