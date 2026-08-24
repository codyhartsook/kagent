package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kagent-dev/kagent/go/core/cli"
	clioutput "github.com/kagent-dev/kagent/go/core/cli/internal/cli/output"
	"github.com/kagent-dev/kagent/go/core/cli/internal/config"
)

func loadConfig() (*config.Config, error) {
	if err := config.Init(); err != nil {
		return nil, err
	}
	return config.Get()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-done
		fmt.Fprintln(os.Stderr, "kagent aborted.")
		fmt.Fprintln(os.Stderr, "Exiting.")
		cancel()
	}()

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	if err := cli.New(cfg).ExecuteContext(ctx); err != nil {
		if cfg.OutputFormat == string(clioutput.Agent) {
			_ = clioutput.WriteError(os.Stdout, err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
