package main

import (
	"context"
	"fmt"
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
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.RegisterTool(polaris.Tool{
		Name:        "web@httpd_log",
		Description: "get httpd access log",
		Parameters: polaris.Object{
			Properties: polaris.Properties{
				"lines": polaris.Int{Description: "number of lines", Required: true},
			},
		},
		Response: polaris.Object{
			Properties: polaris.Properties{
				"log": polaris.String{Description: "log content", Required: true},
			},
		},
		Handler: func(r *polaris.ReqCtx) (polaris.Resp, error) {
			lines := r.Int("lines")
			return polaris.Resp{
				"log": fmt.Sprintf("returning %d lines of httpd log...\n127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] \"GET /apache_pb.gif HTTP/1.0\" 200 2326", lines),
			}, nil
		},
	}); err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log.Println("Web agent (web@httpd_log) registered.")
	<-ctx.Done()
}
