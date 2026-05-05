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
		polaris.RequestTimeout(3*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.RegisterTool(polaris.Tool{
		Name:        "ask_web_agent",
		Description: "delegate tasks to the web expert agent (can fetch logs, check status, etc.)",
		Parameters: polaris.Object{
			Properties: polaris.Properties{
				"task": polaris.String{Description: "description of the task for the web agent", Required: true},
			},
		},
		Response: polaris.Object{
			Properties: polaris.Properties{
				"result": polaris.String{Description: "result from the web agent", Required: true},
			},
		},
		Handler: func(r *polaris.ReqCtx) (polaris.Resp, error) {
			task := r.String("task")
			log.Printf("Router agent received task: %s\n", task)

			ctx := context.Background()
			session, err := conn.Use(
				ctx,
				polaris.UseModel("gemini-3.1-pro-preview"),
				polaris.UseNamespace("web"), // limit to tools in "web@" namespace
			)
			if err != nil {
				return nil, err
			}

			it, err := session.SendText(task)
			if err != nil {
				return nil, err
			}

			result := ""
			for msg, err := range it {
				if err != nil {
					return nil, err
				}
				result += msg
			}
			return polaris.Resp{"result": result}, nil
		},
	}); err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log.Println("Router agent (ask_web_agent) registered.")
	<-ctx.Done()
}
