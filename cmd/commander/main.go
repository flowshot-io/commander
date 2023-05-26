package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/flowshot-io/commander/internal/commander"
	"github.com/flowshot-io/commander/internal/commander/config"
	"github.com/flowshot-io/x/pkg/logger"
)

const AppName = "commander"

func main() {
	configDir := "config"

	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		fmt.Println("Unable to load configuration:", err)
		os.Exit(1)
	}

	logger := logger.New(&logger.Options{Pretty: true})

	commander, err := commander.New(
		commander.WithConfig(cfg),
		commander.WithServices([]string{"frontend", "blenderfarm", "blendernode"}),
		commander.WithLogger(logger),
	)
	if err != nil {
		logger.Error("Unable to create commander", map[string]interface{}{"Error": err.Error()})
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Start the commander in a separate Goroutine
	go func() {
		defer wg.Done()
		err = commander.Start()
		if err != nil {
			logger.Error("Unable to start commander", map[string]interface{}{"Error": err.Error()})
			os.Exit(1)
		}
	}()

	// Listen for os.Interrupt signals (e.g., Ctrl+C)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan
	logger.Info("Received interrupt signal, stopping all services...")
	commander.Stop()

	wg.Wait()
}
