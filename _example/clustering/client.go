package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/octu0/polaris"
	"github.com/pkg/errors"
)

func registerCalculatorAgent(conn *polaris.Conn) error {
	return conn.RegisterTool(polaris.Tool{
		Name:        "calculator",
		Description: "simple calculator",
		Parameters: polaris.Object{
			Description: "calc params",
			Properties: polaris.Properties{
				"operation": polaris.StringEnum{
					Description: "operation to perform",
					Values:      []string{"add", "subtract", "multiply", "divide"},
					Required:    true,
				},
				"a": polaris.Float{
					Description: "first operand",
					Required:    true,
				},
				"b": polaris.Float{
					Description: "second operand",
					Required:    true,
				},
			},
		},
		Response: polaris.Object{
			Description: "result of calculation",
			Properties: polaris.Properties{
				"result": polaris.Float{
					Description: "result",
					Required:    true,
				},
			},
		},
		Handler: func(ctx *polaris.Ctx) error {
			fmt.Println("called handler...")
			operation := ctx.String("operation")
			a := ctx.Float64("a")
			b := ctx.Float64("b")

			result, err := func() (float64, error) {
				switch operation {
				case "add":
					return a + b, nil
				case "subtract":
					return a - b, nil
				case "multiply":
					return a * b, nil
				case "divide":
					if b == 0 {
						return 0, errors.Errorf("divide by zero")
					}
					return a / b, nil
				default:
					return 0, errors.Errorf("unknown operand: %s", operation)
				}
			}()
			if err != nil {
				return errors.WithStack(err)
			}

			ctx.Set(polaris.Resp{
				"result": result,
			})
			return nil
		},
	})
}

func main() {
	conn, err := polaris.Connect(polaris.ConnectAddress("127.0.0.1", "5222"))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := registerCalculatorAgent(conn); err != nil {
		panic(err)
	}
	fmt.Println("client started")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// separate conn
	connClient, err := polaris.Connect(polaris.NatsURL("nats://127.0.0.1:5222", "nats://127.0.0.1:5223", "nats://127.0.0.1:5224"))
	if err != nil {
		panic(err)
	}
	defer connClient.Close()

	resp, err := connClient.Call(ctx, "calculator", polaris.Req{"operation": "add", "a": 1, "b": 2})
	if err != nil {
		panic(err)
	}
	fmt.Println("result = ", resp.Float64("result", -10))
}
