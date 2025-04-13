# polaris clustering

## Registry cluster

Registry can be clustered by using `WithClusterOption` and `WithRoutes`.

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

run first registry process

```shell
$ go run registry.go 1
```

second registry process

```shell
$ go run registry.go 2
```

thrid registry process

```shell
$ go run registry.go 3
```

## Run tool

Tool connects to multiple clusters with the following options:

```go
conn, _ := polaris.Connect(
	polaris.NatsURL(
		"nats://127.0.0.1:4222", "nats://127.0.0.1:4223", "nats://127.0.0.1:4224",
	)
)
conn.RegisterTool(polaris.Tool{Name: "calculator", ...})
```

It allows processing to continue even if the registry going down.

```shell
$ go run tool.go
```

## Using tool

To invoke Tool/Agent, specify multiple clusters with the following options:

```go
client, _ := polaris.Connect(
	polaris.NatsURL(
		"nats://127.0.0.1:4222", "nats://127.0.0.1:4223", "nats://127.0.0.1:4224",
	)
)
client.Call(ctx, "calculator", polaris.Req{...})
```

It allows processing to continue even if the registry going down.

```shell
$ go run client.go
```