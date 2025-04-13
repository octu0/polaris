package main

import (
	"context"
	"fmt"

	"github.com/octu0/polaris"
)

func main() {
	ctx := context.TODO()

	client, err := polaris.Connect(polaris.NatsURL("nats://127.0.0.1:4222", "nats://127.0.0.1:4223", "nats://127.0.0.1:4224"))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	resp, err := client.Call(ctx, "calculator", polaris.Req{"operation": "add", "a": 1, "b": 2})
	if err != nil {
		panic(err)
	}
	fmt.Println("result = ", resp.Float64("result", -10))
}
