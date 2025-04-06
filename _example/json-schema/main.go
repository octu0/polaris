package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/octu0/polaris"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	session, err := polaris.Use(
		ctx,
		polaris.UseModel("gemini-2.5-pro-exp-03-25"),
		polaris.UseSystemInstruction(
			polaris.AddTextSystemInstruction("Output must be in Japanese."),
		),
		polaris.UseTemperature(0.2),
		polaris.UseJSONOutput(polaris.Object{
			Description: "result of each",
			Properties: polaris.Properties{
				"resultA": polaris.Int{
					Description: "result 1",
					Required:    true,
				},
				"resultB": polaris.Int{
					Description: "result 2",
					Required:    true,
				},
			},
		}),
	)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	prompt := `
		execute this task:
		1. Sum 35 and 21
		2. multiply by 88 using one previous answer.

		output(JSON schema):
		ret={
		  "resultA": int,
		  "resultB": int,
		}
		Return: ret
	`
	it, err := session.SendText(prompt)
	if err != nil {
		panic(err)
	}
	for msg, err := range it {
		if err != nil {
			panic(err)
		}
		fmt.Println(msg)
	}
}
