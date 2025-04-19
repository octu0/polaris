package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/octu0/polaris"
)

func main() {
	conn, err := polaris.Connect(
		polaris.ConnectAddress("127.0.0.1", "4222"),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.RegisterTool(polaris.Tool{
		Name:        "calculatorA",
		Description: "calculatorA",
		Parameters: polaris.Object{
			Description: "required two arguments",
			Properties: polaris.Properties{
				"a": polaris.Int{
					Description: "first number",
					Default:     0,
					Required:    true,
				},
				"b": polaris.Int{
					Description: "second number",
					Default:     0,
					Required:    true,
				},
			},
		},
		Response: polaris.Object{
			Description: "response",
			Properties: polaris.Properties{
				"result": polaris.Int{
					Description: "result of calculatorA",
					Required:    true,
				},
			},
		},
		Handler: func(ctx *polaris.Ctx) error {
			log.Println("function calling calculatorA")
			ctx.Set(polaris.Resp{
				"result": ctx.Int("a") + ctx.Int("b"),
			})
			return nil
		},
		ErrorHandler: func(err error) {
			log.Printf("error: %+v", err)
		},
	}); err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log.Println("tool running")
	<-ctx.Done()
}
