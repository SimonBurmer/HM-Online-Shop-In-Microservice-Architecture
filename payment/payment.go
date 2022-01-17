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
	Payments map[uint32]*api.PaymentStorage
	api.UnimplementedPaymentServer
}

func (s *Server) NewPayment(in *api.NewPaymentRequest) {
	log.Printf("received new payment request of: orderId: %v, value: %v", in.GetOrderId(), in.GetValue())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received new payment request of: orderId: %v, value: %v", in.GetOrderId(), in.GetValue()), Subject: "Payment.NewPayment"})
	if err != nil {
		panic(err)
	}

	// Neues Payment-Objekt erstellen und abspeichern
	s.Payments[in.GetOrderId()] = &api.PaymentStorage{OrderId: in.GetOrderId(), Value: in.GetValue(), Canceled: false}

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

	// Payment laden und bezahlen
	out := s.getPayment(in.GetOrderId())
	newValue := out.GetValue() - in.GetValue()
	s.Payments[in.GetOrderId()] = &api.PaymentStorage{OrderId: out.OrderId, Value: newValue, Canceled: false}

	// Überprüfen ob Payment vollständig bezahlt wurde
	if newValue <= 0 {
		log.Printf("successfully payed payment of: orderId: %v", in.GetOrderId())
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("successfully payed payment of: orderId: %v", in.GetOrderId()), Subject: "Payment.PayPayment"})
		if err != nil {
			panic(err)
		}

		// Order-Service informieren, dass Payment vollständig bezahlt wurde
		paymentUpdate := &api.OrderPaymentUpdate{OrderId: in.GetOrderId()}
		err = s.Nats.Publish("order.payment", paymentUpdate)
		if err != nil {
			panic(err)
		}

	} else {
		log.Printf("payment of: orderId: %v has %v value left to pay", in.GetOrderId(), newValue)
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("payment of: orderId: %v has %v value left to pay", in.GetOrderId(), newValue), Subject: "Payment.PayPayment"})
		if err != nil {
			panic(err)
		}
	}
	return &api.PayPaymentReply{OrderId: out.GetOrderId(), StillToPay: out.GetValue()}, nil
}

func (s *Server) CancelPayment(in *api.CancelPaymentRequest) {
	log.Printf("received cancel payment request of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received cancel payment request of: orderId: %v", in.GetOrderId()), Subject: "Payment.CancelPayment"})
	if err != nil {
		panic(err)
	}

	// Payment laden (checken ob Payment mit gegebener Id existiert)
	out := s.getPayment(in.GetOrderId())

	// Bereits bezahlte Summe der Order Zurückerstatten
	if out.GetValue() > 0 {
		log.Printf("completely refound payment of: orderId: %v", in.GetOrderId())
		s.RefundPayment(&api.RefundPaymentRequest{OrderId: out.OrderId, CustomerName: in.GetCustomerName(), CustomerAddress: in.GetCustomerAddress(), Value: out.GetValue()})
	}

	// Payment als storniert markieren
	out.Canceled = true
	s.Payments[in.GetOrderId()] = out

	log.Printf("successfully canceled payment of: orderId: %v", in.GetOrderId())
	err = s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("successfully canceled payment of: orderId: %v", in.GetOrderId()), Subject: "Payment.CancelPayment"})
	if err != nil {
		panic(err)
	}
}

func (s *Server) RefundPayment(in *api.RefundPaymentRequest) {
	log.Printf("received refund payment request of: orderId: %v", in.GetOrderId())
	err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("received refund payment request of: orderId: %v", in.GetOrderId()), Subject: "Payment.RefundPayment"})
	if err != nil {
		panic(err)
	}

	// Payment laden und um Refund-Betrag mindern
	out := s.getPayment(in.GetOrderId())
	newValue := out.GetValue() - in.GetValue()
	s.Payments[in.GetOrderId()] = &api.PaymentStorage{OrderId: out.OrderId, Value: newValue, Canceled: false}

	// Refund-Betrag zurückzahlen
	log.Printf("successfully refund value: %v  to customer %v %v of orderID: %v", in.GetValue(), in.GetCustomerName(), in.GetCustomerAddress(), in.GetOrderId())
	err = s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("successfully refund value: %v  to customer %v %v of orderID: %v", in.GetValue(), in.GetCustomerName(), in.GetCustomerAddress(), in.GetOrderId()), Subject: "Payment.RefundPayment"})
	if err != nil {
		panic(err)
	}
}

func (s *Server) getPayment(orderID uint32) *api.PaymentStorage {
	out, ok := s.Payments[orderID]
	if !ok {
		err := s.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no payment with orderId: %v", orderID), Subject: "Payment.getPayment"})
		if err != nil {
			panic(err)
		}
		log.Fatalf("no payment with orderId: %v", orderID)
	}

	return out
}
