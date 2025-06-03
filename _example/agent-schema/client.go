package main

import (
	"context"
	"fmt"
	"time"

	"github.com/octu0/polaris"
)

func main() {
	conn, err := polaris.Connect(
		polaris.ConnectAddress("127.0.0.1", "4222"),
		polaris.ConnectTimeout(3*time.Second),
		polaris.RequestTimeout(10*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ctx := context.TODO()
	session, err := conn.Use(
		ctx,
		polaris.UseModel("gemini-2.5-pro-preview-05-06"),
		polaris.UseSystemInstruction(
			polaris.AddTextSystemInstruction("Output must be in Japanese."),
		),
		polaris.UseTemperature(0.2),
	)
	if err != nil {
		panic(err)
	}

	prompt := `
		tell me today condition.
		my fortune and tokyo weather.
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
