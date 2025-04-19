package main

import (
	"context"
	"fmt"
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

	ctx := context.TODO()
	session, err := conn.Use(
		ctx,
		polaris.UseModel("gemini-2.5-pro-exp-03-25"),
		polaris.UseSystemInstruction(
			polaris.AddTextSystemInstruction("Output must be in Japanese."),
		),
		polaris.UseTemperature(0.2),
	)
	if err != nil {
		panic(err)
	}

	prompt := `
		execute this task:
		1. Use 35 and 21 for calculatorA 
		2. Use result of calculatorA and 88 for calculatorB 

		tell me each result
	`
	it, err := session.SendText(prompt)
	if err != nil {
		panic(err))
	}
	for msg, err := range it {
		if err != nil {
			panic(err)
		}
		fmt.Println(msg)
	}
}
