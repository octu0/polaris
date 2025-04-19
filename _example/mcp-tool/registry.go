package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/octu0/polaris"
)

func main() {
	registry, err := polaris.CreateRegistry(
		polaris.WithBind("127.0.0.1", 4222),
	)
	if err != nil {
		panic(err)
	}
	defer registry.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	fmt.Println("registry started.")
	<-ctx.Done()
}
