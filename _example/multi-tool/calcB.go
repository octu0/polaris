package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/octu0/polaris"
)

func main() {
	conn, err := polaris.Connect(
		polaris.ConnectAddress("127.0.0.1", "4222"),
		polaris.ConnectTimeout(3*time.Second),
		polaris.AllowReconnect(true),
		polaris.MaxReconnects(-1),
		polaris.ReconnectWait(5*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.RegisterTool(polaris.Tool{
		Name:        "calculatorB",
		Description: "calculatorB",
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
					Description: "result of calculatorB",
					Required:    true,
				},
			},
		},
		Handler: func(ctx *polaris.Ctx) error {
			log.Println("function calling calculatorB")
			a := ctx.Int("a")
			b := ctx.Int("b")
			ctx.Set(polaris.Resp{
				"result": a * b,
			})
			return nil
		},
	}); err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log.Println("calculatorB running")
	<-ctx.Done()
}
