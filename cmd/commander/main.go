package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/flowshot-io/commander/internal/commander"
	"github.com/flowshot-io/commander/internal/commander/config"
)

const AppName = "commander"

func main() {
	configDir := "config"

	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		fmt.Println("Unable to load configuration:", err)
		os.Exit(1)
	}

	logger := log.Default()

	commander, err := commander.New(
		commander.WithConfig(cfg),
		commander.WithServices([]string{"frontend", "blenderfarm", "blendernode"}),
		commander.WithLogger(logger),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Start the commander in a separate Goroutine
	go func() {
		defer wg.Done()
		commander.Start()
	}()

	// Listen for os.Interrupt signals (e.g., Ctrl+C)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan
	fmt.Println("Received interrupt signal, stopping all services...")
	commander.Stop()

	wg.Wait()
}
