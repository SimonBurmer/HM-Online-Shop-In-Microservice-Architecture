package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	// Verbindung zu Nats aufgaben
	nc, err := nats.Connect("nats:4222")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to nats")
	}
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create NewEncodedConn")
	}
	defer c.Close()

	// Nats Channel subscriben
	subscription, err := c.Subscribe("log", func(msg *api.Log) {
		log.Info().
			Str("subj", msg.Subject).
			Str("data", string(msg.Message)).
			Msg("received")
	})
	if err != nil {
		log.Fatal().Err(err).Msg("cannot subscribe")
	}
	defer subscription.Unsubscribe()

	var wc sync.WaitGroup
	wc.Add(1)
	wc.Wait()
}
