# polaris registry Clustering

This example demonstrates how to set up a clustered Polaris registry for high availability and fault tolerance. By running multiple registry servers and configuring clients and tools to connect to all of them, you can ensure that your system continues to function even if some registry servers become unavailable.

## Overview

In production environments, having a single point of failure can lead to system-wide outages. polaris addresses this issue by supporting registry server clustering, which provides:

- **High Availability**: If one registry server fails, others in the cluster can continue to handle requests
- **Load Balancing**: Distribute the load across multiple registry servers
- **Fault Tolerance**: The system remains operational even when some components fail
- **Seamless Failover**: Clients and tools automatically connect to available registry servers

## Prerequisites

Before running this example, you need to set up the following environment variables for Google Cloud Platform authentication:

```bash
export GOOGLE_CLOUD_PROJECT=your_project_id
export GOOGLE_CLOUD_LOCATION=your_gcp_project_location
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credential.json
```

These environment variables are required to authenticate with Google Cloud and access the AI models used in polaris.

## Registry Cluster Setup

Registry servers can be clustered by using `WithClusterOption` and `WithRoutes`. Each registry server needs to be configured with:

1. A unique binding address and port
2. Cluster configuration options
3. Routes to other registry servers (for the second and subsequent servers)

```go
registry, _ := polaris.CreateRegistry(
	polaris.WithBind("127.0.0.1", 4223),
	polaris.WithClusterOption(
		polaris.WithClusterName("my-cluster"),
		polaris.WithClusterHost("127.0.0.1"),
		polaris.WithClusterPort(5223),
		polaris.WithClussterAdvertise("127.0.0.1:4223"),
	),
	polaris.WithRoutes("nats://127.0.0.1:5222"), // this option is used for the second and later units.
)
```

### Cluster Configuration Options

- `WithClusterName`: Sets a name for the cluster (all nodes should use the same name)
- `WithClusterHost`: Specifies the host for cluster communication
- `WithClusterPort`: Specifies the port for cluster communication
- `WithClussterAdvertise`: Advertises the address that other servers can use to connect to this server
- `WithRoutes`: Specifies the addresses of other servers in the cluster (only needed for the second and subsequent servers)

### Running Multiple Registry Servers

To start a cluster, you need to run multiple registry processes with different configurations:

First registry process:
```shell
$ go run registry.go 1
```

Second registry process:
```shell
$ go run registry.go 2
```

Third registry process:
```shell
$ go run registry.go 3
```

The example code automatically assigns different ports based on the node ID:
- Node 1: Port 4222 (client), 5222 (cluster)
- Node 2: Port 4223 (client), 5223 (cluster)
- Node 3: Port 4224 (client), 5224 (cluster)

## Connecting Tools to Clustered Registry

Tools can connect to multiple registry servers by specifying all server addresses in the `NatsURL` option:

```go
conn, _ := polaris.Connect(
	polaris.NatsURL(
		"nats://127.0.0.1:4222", "nats://127.0.0.1:4223", "nats://127.0.0.1:4224",
	)
)
conn.RegisterTool(polaris.Tool{Name: "calculator", ...})
```

This configuration allows the tool to:
- Connect to any available registry server
- Automatically reconnect to another server if the current one fails
- Continue processing requests without interruption

To run the example tool:
```shell
$ go run tool.go
```

## Using Tools with Clustered Registry

Clients follow the same pattern, specifying all registry server addresses:

```go
client, _ := polaris.Connect(
	polaris.NatsURL(
		"nats://127.0.0.1:4222", "nats://127.0.0.1:4223", "nats://127.0.0.1:4224",
	)
)
client.Call(ctx, "calculator", polaris.Req{...})
```

This ensures that clients can:
- Discover and use tools registered on any registry server in the cluster
- Continue operating even if some registry servers become unavailable
- Automatically reconnect to available servers without manual intervention

To run the example client:
```shell
$ go run client.go
```

## How Failover Works

The clustering mechanism works as follows:

1. Registry servers form a mesh network where each server is aware of others
2. When a tool registers with one registry server, the registration is propagated to all servers in the cluster
3. Clients and tools maintain connections to multiple registry servers
4. If a registry server fails:
   - Connected clients and tools detect the failure
   - They automatically switch to another available server
   - Operations continue without interruption

This provides seamless failover and ensures that your system remains operational even when individual components fail.

## Testing Failover

To test the failover capability:

1. Start all three registry servers
2. Start the tool and client
3. Terminate one of the registry servers (e.g., press Ctrl+C in its terminal)
4. Observe that the tool and client continue to function normally

This demonstrates the high availability and fault tolerance provided by the Polaris clustering feature.
