package main

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"golang.org/x/sys/unix"

	"pacabid/internal/broker"
	"pacabid/internal/strategy"
)

type pacabid struct {
	client broker.Client
	quit   chan struct{}
}

func main() {
	client := broker.NewAlpaca()
	strat := strategy.NewMeanReversion("T", 25)
	strat.Prepare(0, client)
	quit := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := strat.Run(quit); err != nil {
			panic(err)
		}
	}()
	log.Println("pacabid is running.")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGINT, unix.SIGTERM)
	<-signals
	log.Println("Stopping pacabid.")
	close(quit)
	wg.Wait()
	log.Println("Exiting positions.")
	if err := client.ExitAllPositions(); err != nil {
		log.Fatal("Failed to exit all positions: ", err)
	}
}
