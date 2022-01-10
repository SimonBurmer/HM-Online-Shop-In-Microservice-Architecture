package main

import (
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/client"
)

func main() {
	// Verbindung zu Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // 127.0.0.1 / host.docker.internal
		Password: "",           // no password set
	})

	// Verbindung zu NATS
	nc, err := nats.Connect("nats:4222") // 127.0.0.1
	if err != nil {
		log.Fatal(err)
	}
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	client := client.Client{Nats: c, Redis: rdb}
	log.Printf("Client object successfully created")

	client.Scenarios(os.Getenv("scenario"))

	log.Printf("Client successfully finished")
}
