package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/octu0/polaris"
)

func main() {
	nodeID := "1"
	if 1 < len(os.Args) {
		nodeID = os.Args[1]
	}

	id, _ := strconv.Atoi(nodeID)
	port := 4222 + (id - 1)
	clusterPort := 5222 + (id - 1)

	routes := ""
	if 1 < id {
		routes = "nats://127.0.0.1:5222"
	}

	registry, err := polaris.CreateRegistry(
		polaris.WithBind("127.0.0.1", port),
		polaris.WithClusterOption(
			polaris.WithClusterName("polaris-cluster"),
			polaris.WithClusterHost("127.0.0.1"),
			polaris.WithClusterPort(clusterPort),
			polaris.WithClussterAdvertise(fmt.Sprintf("127.0.0.1:%d", port)),
		),
		polaris.WithRoutes(routes),
	)
	if err != nil {
		panic(err)
	}
	defer registry.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	fmt.Printf("Registry node %s started on port %d (cluster port: %d)\n", nodeID, port, clusterPort)

	<-ctx.Done()
	fmt.Printf("Registry node %s shutting down...\n", nodeID)
}
