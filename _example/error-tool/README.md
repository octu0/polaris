# Error Handling in polaris Tools

This example demonstrates how to handle errors in polaris tools and how agents can receive error information. It showcases the error propagation mechanism and the use of `ErrorHandler` to process errors when they occur.

## Overview

In real-world applications, tools may encounter errors during execution. polaris provides a robust error handling mechanism that allows:

1. Tools to return errors that are properly propagated to agents
2. Custom error handling logic through the `ErrorHandler` function
3. Agents to receive and process error information appropriately

This approach ensures that errors are handled gracefully and that the system can respond appropriately when things go wrong.

## Prerequisites

Before running this example, you need to set up the following environment variables for Google Cloud Platform authentication:

```bash
export GOOGLE_CLOUD_PROJECT=your_project_id
export GOOGLE_CLOUD_LOCATION=your_gcp_project_location
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credential.json
```

These environment variables are required to authenticate with Google Cloud and access the Gemini AI model used in this example.

## Components

This example consists of three main components:

### 1. Registry Server (`registry.go`)

The registry server acts as a central hub for registering and managing tools. It binds to a local address and port, allowing clients and tools to connect.

```go
registry, err := polaris.CreateRegistry(
    polaris.WithBind("127.0.0.1", 4222),
)
```

### 2. Tool Implementation with Error Handling (`tool.go`)

The tool implementation defines a calculator tool that intentionally returns an error for demonstration purposes. It also includes an `ErrorHandler` to process the error:

```go
conn.RegisterTool(polaris.Tool{
    Name:        "calculator",
    Description: "calculator",
    Parameters: polaris.Object{
        // Parameter definitions...
    },
    Response: polaris.Object{
        // Response definition...
    },
    Handler: func(r *polaris.ReqCtx) (polaris.Resp, error) {
        return nil, fmt.Errorf("!!!this function does not support!!! args=%v", r.Req())
    },
    ErrorHandler: func(err error) {
        log.Printf("error: %+v", err)
    },
})
```

### 3. Client Application (`client.go`)

The client application connects to the registry, creates a session with the AI model, and sends a prompt that attempts to use the calculator tool:

```go
prompt := `
    execute this task:
    1. Use 35 and 21 for calculator
    2. Use 100 and 23 for calculator

    tell me each result.
`
```

## Key Features

### Error Propagation

When a tool's `Handler` function returns an error, polaris automatically propagates this error information to the agent. This allows the agent to be aware of the error and respond accordingly.

### Custom Error Handling with `ErrorHandler`

The `ErrorHandler` function provides a way to define custom error handling logic that is executed when an error occurs in the tool's `Handler` function:

```go
ErrorHandler: func(err error) {
    log.Printf("error: %+v", err)
},
```

This allows for various error handling strategies such as:
- Logging the error for debugging
- Attempting recovery or fallback operations
- Notifying administrators or monitoring systems
- Collecting error statistics

### Agent Error Awareness

When an error occurs in a tool, the agent receives this information and can adapt its response accordingly. This creates a more robust interaction where the agent can acknowledge the error and suggest alternatives or provide appropriate guidance to the user.

## Usage

To run this example:

1. Set up the required environment variables as described in the Prerequisites section.

2. Start the registry server:
```shell
$ go run registry.go
```

3. In a separate terminal, start the tool:
```shell
$ go run tool.go
```

4. In another terminal, run the client:
```shell
$ go run client.go
```

## Expected Behavior

When you run this example:

1. The client sends a prompt asking to use the calculator tool with specific inputs
2. The tool's `Handler` function returns an error
3. The `ErrorHandler` function logs the error
4. The agent receives the error information
5. The agent responds to the user, acknowledging the error and potentially suggesting alternatives

This demonstrates how errors are propagated through the system and how both tools and agents can handle them appropriately.

## Benefits of polaris Error Handling

The error handling mechanism in polaris provides several benefits:

1. **Robustness**: The system can handle unexpected situations gracefully
2. **Transparency**: Errors are clearly communicated to agents and users
3. **Flexibility**: Custom error handling logic can be implemented for each tool
4. **Debugging**: Detailed error information helps identify and fix issues

This approach ensures that your polaris-based applications can handle errors in a way that maintains a good user experience even when things don't go as planned.
