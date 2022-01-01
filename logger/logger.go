package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	nc, err := nats.Connect("host.docker.internal:4222")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to nats")
	}
	defer nc.Close()

	subscription, err := nc.Subscribe("log.*", func(msg *nats.Msg) {
		log.Info().
			Str("subj", msg.Subject).
			Str("data", string(msg.Data)).
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
