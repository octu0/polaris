package main

import (
	"context"
	"fmt"
	"log"
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
		polaris.RequestTimeout(3*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ctx := context.TODO()
	// Call without namespace, it will only see "ask_web_agent" not "web@httpd_log"
	session, err := conn.Use(
		ctx,
		polaris.UseModel("gemini-3.1-pro-preview"),
		polaris.UseSystemInstruction(
			polaris.AddTextSystemInstruction("You are a helpful orchestrator. You can use available tools."),
		),
	)
	if err != nil {
		panic(err)
	}

	prompt := "Please check the web server logs and get the last 5 lines."
	log.Printf("User: %s\n", prompt)

	it, err := session.SendText(prompt)
	if err != nil {
		panic(err)
	}

	log.Println("AI Response:")
	for msg, err := range it {
		if err != nil {
			panic(err)
		}
		fmt.Println(msg)
	}
}
