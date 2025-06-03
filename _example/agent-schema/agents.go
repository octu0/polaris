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
		Handler: func(r *polaris.ReqCtx) (polaris.Resp, error) {
			log.Printf("function call: %s", toolName)
			myTool, _ := conn.Tool(toolName)
			gen, err := polaris.GenerateJSON(
				ctx,
				polaris.UseModel("gemini-2.5-flash-preview-05-20"),
				polaris.UseSystemInstruction(
					polaris.AddTextSystemInstruction("Output must be in Japanese."),
				),
				polaris.UseJSONOutput(myTool.Response),
				polaris.UseTemperature(0.5),
			)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			currDate, err := conn.Call(ctx, "getCurrentDate", polaris.Req{})
			if err != nil {
				return nil, errors.WithStack(err)
			}

			prompt := fmt.Sprintf(`
				Specify the City name and Month, making the weather information

				City:
				%s
				Month:
				%s
			`, r.String("cityName"), currDate.String("month", ""))

			resp, err := gen(prompt)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return resp, nil
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
		Handler: func(r *polaris.ReqCtx) (polaris.Resp, error) {
			log.Printf("function call: %s", toolName)
			t, _ := conn.Tool(toolName)
			gen, err := polaris.GenerateJSON(
				ctx,
				polaris.UseModel("gemini-2.5-flash-preview-05-20"),
				polaris.UseSystemInstruction(
					polaris.AddTextSystemInstruction("Output must be in Japanese."),
				),
				polaris.UseJSONOutput(t.Response),
				polaris.UseTemperature(0.5),
			)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			resp, err := gen(`
				Perform a simple, omikuji-style fortune telling for today.
				Give a rndom luck level ('Great luck', 'Good luck', 'Small luck') and a short positive message.
				Make it clear this is just for fun from AI.
			`)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return resp, nil
		},
	})
}

func registerCurrentDate(ctx context.Context, conn *polaris.Conn) error {
	return conn.RegisterTool(polaris.Tool{
		Name:        "getCurrentDate",
		Description: "get current date",
		Parameters:  polaris.Object{},
		Response: polaris.Object{
			Properties: polaris.Properties{
				"year": polaris.String{
					Description: "current year",
					Required:    true,
				},
				"month": polaris.String{
					Description: "current month",
					Required:    true,
				},
				"day": polaris.String{
					Description: "current day",
					Required:    true,
				},
			},
		},
		Handler: func(r *polaris.ReqCtx) (polaris.Resp, error) {
			log.Printf("function call: getCurrentDate")
			now := time.Now()
			return polaris.Resp{
				"year":  fmt.Sprintf("%d", now.Year()),
				"month": now.Month().String(),
				"day":   fmt.Sprintf("%d", now.Day()),
			}, nil
		},
	})
}

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
	if err := registerCurrentDate(ctx, conn); err != nil {
		panic(err)
	}

	<-ctx.Done()
}
