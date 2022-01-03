package payment

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats      *nats.Conn
	Payments  map[uint32]*api.NewPaymentRequest
	PaymentID uint32
	api.UnimplementedPaymentServer
}

func (s Server) NewPayment(ctx context.Context, in *api.NewPaymentRequest) (*api.PaymentReply, error) {
	log.Printf("received new payment request of: orderID: %v, value: %v", in.GetOrderId(), in.GetValue())
	err := s.Nats.Publish("log.payment", []byte(fmt.Sprintf("received new payment request of: orderID: %v, value: %v", in.GetOrderId(), in.GetValue())))
	if err != nil {
		panic(err)
	}

	s.PaymentID++
	s.Payments[s.PaymentID] = in

	log.Printf("successfully created new payment: orderID: %v, value: %v", in.GetOrderId(), in.GetValue())
	err = s.Nats.Publish("log.payment", []byte(fmt.Sprintf("successfully created new payment: orderID: %v, value: %v", in.GetOrderId(), in.GetValue())))
	if err != nil {
		panic(err)
	}

	return &api.PaymentReply{Id: s.PaymentID, OrderId: in.GetOrderId(), Value: in.GetValue()}, nil
}

func (s Server) GetPayment(ctx context.Context, in *api.GetPaymentRequest) (*api.PaymentReply, error) {
	log.Printf("received get payment request of: ID: %v", in.GetId())
	err := s.Nats.Publish("log.payment", []byte(fmt.Sprintf("received get payment request of: ID: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	out, ok := s.Payments[in.GetId()]
	if !ok {
		// Fehlerbehandlung!!
		err = s.Nats.Publish("log.payment", []byte(fmt.Sprintf("no payment with id: %v", in.GetId())))
		if err != nil {
			panic(err)
		}
		log.Fatalf("no payment with id: %v", in.GetId())
	}

	log.Printf("successfully loaded payment of: id: %v", in.GetId())
	err = s.Nats.Publish("log.payment", []byte(fmt.Sprintf("successfully loaded payment of: id: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	return &api.PaymentReply{Id: s.PaymentID, OrderId: out.GetOrderId(), Value: out.GetValue()}, nil
}

func (s Server) DeletePayment(ctx context.Context, in *api.DeletePaymentRequest) (*api.PaymentReply, error) {
	log.Printf("received delete payment request of: ID: %v", in.GetId())
	err := s.Nats.Publish("log.payment", []byte(fmt.Sprintf("received delete payment request of: ID: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	out, ok := s.Payments[in.GetId()]
	if !ok {
		// Fehlerbehandlung!!
		err = s.Nats.Publish("log.payment", []byte(fmt.Sprintf("no payment with id: %v", in.GetId())))
		if err != nil {
			panic(err)
		}
		log.Fatalf("no payment with id: %v", in.GetId())
	}

	delete(s.Payments, in.GetId())

	log.Printf("successfully deleted payment of: id: %v", in.GetId())
	err = s.Nats.Publish("log.payment", []byte(fmt.Sprintf("successfully deleted payment of: id: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	return &api.PaymentReply{Id: s.PaymentID, OrderId: out.GetOrderId(), Value: out.GetValue()}, nil
}

func (s Server) PayPayment(ctx context.Context, in *api.PayPaymentRequest) (*api.PaymentReply, error) {
	log.Printf("received pay payment request of: ID: %v", in.GetId())
	err := s.Nats.Publish("log.payment", []byte(fmt.Sprintf("received pay payment request of: ID: %v", in.GetId())))
	if err != nil {
		panic(err)
	}

	out, ok := s.Payments[in.GetId()]
	if !ok {
		// Fehlerbehandlung!!
		err = s.Nats.Publish("log.payment", []byte(fmt.Sprintf("no payment with id: %v", in.GetId())))
		if err != nil {
			panic(err)
		}
		log.Fatalf("no payment with id: %v", in.GetId())
	}

	newValue := out.GetValue() - in.GetValue()
	s.Payments[in.GetId()] = &api.NewPaymentRequest{OrderId: out.OrderId, Value: newValue}

	if newValue <= 0 {
		log.Printf("successfully payed payment of: id: %v", in.GetId())
		err = s.Nats.Publish("log.payment", []byte(fmt.Sprintf("successfully payed payment of: id: %v", in.GetId())))
		if err != nil {
			panic(err)
		}

		// !!! Bescheid geben Orderservice!!!!!

	} else {
		log.Printf("payment of: id: %v has %v value left to pay", in.GetId(), newValue)
		err = s.Nats.Publish("log.payment", []byte(fmt.Sprintf("payment of: id: %v has %v value left to pay", in.GetId(), newValue)))
		if err != nil {
			panic(err)
		}
	}

	return &api.PaymentReply{Id: s.PaymentID, OrderId: out.GetOrderId(), Value: out.GetValue()}, nil
}
