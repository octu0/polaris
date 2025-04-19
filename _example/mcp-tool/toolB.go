package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

	mcpServer := server.NewMCPServer(
		"test-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	mcpServer.AddTool(mcp.NewTool(
		"calculatorB",
		mcp.WithDescription("calculatorB"),
		mcp.WithNumber("a",
			mcp.Description("First number"),
			mcp.Required(),
		),
		mcp.WithNumber("b",
			mcp.Description("Second number"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("function calling calculatorB: args=%v", request.Params.Arguments)
		arguments := request.Params.Arguments
		a, ok1 := arguments["a"].(float64)
		b, ok2 := arguments["b"].(float64)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("invalid number arguments")
		}
		sum := a + b
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("The sum of %f and %f is %f.", a, b, sum),
				},
			},
		}, nil
	})

	sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL("http://127.0.0.1:8080"))

	wait := make(chan struct{})
	go func() {
		close(wait)
		sseServer.Start("0.0.0.0:8080")
	}()

	<-wait
	initReq := mcp.InitializeRequest{}
	if err := conn.RegisterSSEMCPTools("http://127.0.0.1:8080/sse", initReq); err != nil {
		log.Fatalf("error: %+v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log.Println("tool running")
	<-ctx.Done()
}
