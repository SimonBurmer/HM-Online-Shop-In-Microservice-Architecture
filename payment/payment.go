package payment

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

type Server struct {
	Nats     *nats.EncodedConn
	Payments map[uint32]*api.NewPaymentRequest
	api.UnimplementedPaymentServer
}

func (s *Server) NewPayment(in *api.NewPaymentRequest) {
	log.Printf("received new payment request of: orderId: %v, value: %v", in.GetOrderId(), in.GetValue())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received new payment request of: orderId: %v, value: %v", in.GetOrderId(), in.GetValue()), Subject: "Payment.NewPayment"})
	if err != nil {
		panic(err)
	}

	// Hier evtl. überprüfen ob OrderId existiert!

	s.Payments[in.GetOrderId()] = in

	log.Printf("successfully created new payment: orderId: %v, value: %v", in.GetOrderId(), in.GetValue())
	err = s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("successfully created new payment: orderId: %v, value: %v", in.GetOrderId(), in.GetValue()), Subject: "Payment.NewPayment"})
	if err != nil {
		panic(err)
	}
}

func (s *Server) PayPayment(ctx context.Context, in *api.PayPaymentRequest) (*api.PayPaymentReply, error) {
	log.Printf("received pay payment request of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received pay payment request of: orderId: %v", in.GetOrderId()), Subject: "Payment.PayPayment"})
	if err != nil {
		panic(err)
	}

	out, ok := s.Payments[in.GetOrderId()]
	if !ok {
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no payment with orderId: %v", in.GetOrderId()), Subject: "Payment.PayPayment"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("no payment with orderId: %v", in.GetOrderId())
	}

	newValue := out.GetValue() - in.GetValue()
	s.Payments[in.GetOrderId()] = &api.NewPaymentRequest{OrderId: out.OrderId, Value: newValue}

	if newValue <= 0 {
		log.Printf("successfully payed payment of: orderId: %v", in.GetOrderId())
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("successfully payed payment of: orderId: %v", in.GetOrderId()), Subject: "Payment.PayPayment"})
		if err != nil {
			panic(err)
		}

		// !!! Orderservice bescheidgeben !!!!!

	} else {
		log.Printf("payment of: orderId: %v has %v value left to pay", in.GetOrderId(), newValue)
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("payment of: orderId: %v has %v value left to pay", in.GetOrderId(), newValue), Subject: "Payment.PayPayment"})
		if err != nil {
			panic(err)
		}
	}
	return &api.PayPaymentReply{OrderId: out.GetOrderId(), Value: out.GetValue()}, nil
}

func (s *Server) RefundPayment(in *api.RefundPaymentRequest) {
	log.Printf("received refound payment request of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received refound payment request of: orderId: %v", in.GetOrderId()), Subject: "Payment.RefundPayment"})
	if err != nil {
		panic(err)
	}

	out, ok := s.Payments[in.GetOrderId()]
	if !ok {
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no payment with orderId: %v", in.GetOrderId()), Subject: "Payment.RefundPayment"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("no payment with orderId: %v", in.GetOrderId())
	}

	newValue := out.GetValue() - in.GetValue()
	s.Payments[in.GetOrderId()] = &api.NewPaymentRequest{OrderId: out.OrderId, Value: newValue}

	log.Printf("AUSZAHLUNG PASST!!!")
}
