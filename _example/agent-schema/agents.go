package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/octu0/polaris"
	"github.com/pkg/errors"
)

func main() {
	conn, err := polaris.Connect(polaris.ConnectAddress("127.0.0.1", "4222"))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := registerWeatherAgent(ctx, conn); err != nil {
		panic(err)
	}
	if err := registerFortuneAgent(ctx, conn); err != nil {
		panic(err)
	}

	<-ctx.Done()
}

func registerWeatherAgent(ctx context.Context, conn *polaris.Conn) error {
	toolName := "getWeather"
	return conn.RegisterTool(polaris.Tool{
		Name:        toolName,
		Description: "get weather by city",
		Parameters: polaris.Object{
			Properties: polaris.Properties{
				"cityName": polaris.String{
					Description: "cityName",
					Default:     "tokyo",
					Required:    true,
				},
			},
		},
		Response: polaris.Object{
			Properties: polaris.Properties{
				"temperature": polaris.Int{
					Description: "estimated maximum temperatures",
					Required:    true,
				},
				"sky_condition": polaris.String{
					Description: "sky condition",
					Required:    true,
				},
			},
		},
		Handler: func(c *polaris.Ctx) error {
			log.Printf("function call: %s", toolName)
			t, _ := conn.Tool(toolName)
			gen, err := polaris.GenerateJSON(
				ctx,
				polaris.UseModel("gemini-2.5-pro-exp-03-25"),
				polaris.UseSystemInstruction(
					polaris.AddTextSystemInstruction("Output must be in Japanese."),
				),
				polaris.UseJSONOutput(t.Response),
				polaris.UseTemperature(0.5),
			)
			if err != nil {
				return errors.WithStack(err)
			}

			prompt := fmt.Sprintf(`
				Specify the City name and Month, making the weather information

				City:
				%s
				Month:
				%s
			`, c.String("cityName"), time.Now().Month().String())

			resp, err := gen(prompt)
			if err != nil {
				return errors.WithStack(err)
			}

			c.Set(resp)
			return nil
		},
	})
}

func registerFortuneAgent(ctx context.Context, conn *polaris.Conn) error {
	toolName := "getFortune"
	return conn.RegisterTool(polaris.Tool{
		Name:        toolName,
		Description: "get fortune",
		Parameters:  polaris.Object{},
		Response: polaris.Object{
			Properties: polaris.Properties{
				"result": polaris.String{
					Description: "result",
					Required:    true,
				},
			},
		},
		Handler: func(c *polaris.Ctx) error {
			log.Printf("function call: %s", toolName)
			t, _ := conn.Tool(toolName)
			gen, err := polaris.GenerateJSON(
				ctx,
				polaris.UseModel("gemini-2.5-pro-exp-03-25"),
				polaris.UseSystemInstruction(
					polaris.AddTextSystemInstruction("Output must be in Japanese."),
				),
				polaris.UseJSONOutput(t.Response),
				polaris.UseTemperature(0.5),
			)
			if err != nil {
				return errors.WithStack(err)
			}
			resp, err := gen(`
				Perform a simple, omikuji-style fortune telling for today.
				Give a rndom luck level ('Great luck', 'Good luck', 'Small luck') and a short positive message.
				Make it clear this is just for fun from AI.
			`)
			if err != nil {
				return errors.WithStack(err)
			}

			c.Set(resp)
			return nil
		},
	})
}
