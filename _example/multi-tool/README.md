# Multi-Tool Execution in polaris

This example demonstrates how to use multiple tools in sequence with polaris, allowing AI agents to chain operations and use the output of one tool as input to another.

## Overview

In real-world applications, complex tasks often require multiple operations that build on each other's results. polaris makes it easy for AI agents to:

1. Call multiple tools in sequence
2. Use the output from one tool as input to another
3. Maintain context across tool invocations
4. Create workflows that combine different capabilities

This example showcases a simple calculation workflow where the result of one calculation becomes input to another.

## Prerequisites

Before running this example, you need to set up the following environment variables for Google Cloud Platform authentication:

```bash
export GOOGLE_CLOUD_PROJECT=your_project_id
export GOOGLE_CLOUD_LOCATION=your_gcp_project_location
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credential.json
```

These environment variables are required to authenticate with Google Cloud and access the Gemini AI model used in this example.

## Components

This example consists of four main components:

### 1. Registry Server (`registry.go`)

The registry server acts as a central hub for registering and managing tools. It binds to a local address and port, allowing clients and tools to connect.

```go
registry, err := polaris.CreateRegistry(
    polaris.WithBind("127.0.0.1", 4222),
)
```

### 2. First Calculator Tool (`calcA.go`)

The first calculator tool (`calculatorA`) takes two numbers and returns their sum:

```go
conn.RegisterTool(polaris.Tool{
    Name:        "calculatorA",
    Description: "calculatorA",
    Parameters: polaris.Object{
        Description: "required two arguments",
        Properties: polaris.Properties{
            "a": polaris.Int{
                Description: "first number",
                Default:     0,
                Required:    true,
            },
            "b": polaris.Int{
                Description: "second number",
                Default:     0,
                Required:    true,
            },
        },
    },
    Response: polaris.Object{
        Description: "response",
        Properties: polaris.Properties{
            "result": polaris.Int{
                Description: "result of calculatorA",
                Required:    true,
            },
        },
    },
    Handler: func(ctx *polaris.Ctx) error {
        a := ctx.Int("a")
        b := ctx.Int("b")
        ctx.Set(polaris.Resp{
            "result": a + b,
        })
        return nil
    },
})
```

### 3. Second Calculator Tool (`calcB.go`)

The second calculator tool (`calculatorB`) takes two numbers and returns their product:

```go
conn.RegisterTool(polaris.Tool{
    Name:        "calculatorB",
    Description: "calculatorB",
    Parameters: polaris.Object{
        Description: "required two arguments",
        Properties: polaris.Properties{
            "a": polaris.Int{
                Description: "first number",
                Default:     0,
                Required:    true,
            },
            "b": polaris.Int{
                Description: "second number",
                Default:     0,
                Required:    true,
            },
        },
    },
    Response: polaris.Object{
        Description: "response",
        Properties: polaris.Properties{
            "result": polaris.Int{
                Description: "result of calculatorB",
                Required:    true,
            },
        },
    },
    Handler: func(ctx *polaris.Ctx) error {
        a := ctx.Int("a")
        b := ctx.Int("b")
        ctx.Set(polaris.Resp{
            "result": a * b,
        })
        return nil
    },
})
```

### 4. Client Application (`main.go`)

The client application connects to the registry, creates a session with the AI model, and sends a prompt that instructs the agent to use both tools in sequence:

```go
prompt := `
    execute this task:
    1. Use 35 and 21 for calculatorA 
    2. Use result of calculatorA and 88 for calculatorB 

    tell me each result
`
```

## How Multi-Tool Execution Works

When the AI agent processes the prompt, it:

1. Recognizes the need to call `calculatorA` with inputs 35 and 21
2. Calls `calculatorA` and receives the result (56)
3. Understands it needs to use this result as input to `calculatorB`
4. Calls `calculatorB` with inputs 56 and 88
5. Receives the final result (4,928)
6. Formats a response that includes both calculation results

This demonstrates the agent's ability to:
- Understand multi-step instructions
- Maintain context between tool calls
- Use the output of one tool as input to another
- Present a coherent summary of the entire process

## Usage

To run this example:

1. Set up the required environment variables as described in the Prerequisites section.

2. Start the registry server:
```shell
$ go run registry.go
```

3. In separate terminals, start both calculator tools:

```shell
$ go run calcA.go
```

```shell
$ go run calcB.go
```

4. In another terminal, run the client:

```shell
$ go run main.go
```

## Expected Output

When you run this example, the client will output a response in Japanese (as specified in the system instructions) that explains:

1. The first calculation: 35 + 21 = 56 using calculatorA
2. The second calculation: 56 x 88 = 4928 using calculatorB
3. A summary of both results

The exact wording will vary based on the AI model's generation, but the numerical results should be consistent.

## Benefits of Multi-Tool Execution

The ability to chain multiple tools together provides several benefits:

1. **Complex Workflows**: Enables the creation of sophisticated multi-step processes
2. **Data Flow**: Allows data to flow naturally between different operations
3. **Specialization**: Each tool can focus on a specific task while the agent handles coordination
4. **Flexibility**: New workflows can be created simply by changing the prompt, without modifying the tools

## Real-World Applications

This pattern of chaining multiple tools can be applied to many real-world scenarios:

- Data processing pipelines where data is transformed in multiple stages
- Multi-step business processes that require different operations
- Complex decision trees where each step depends on previous results
- Workflows that combine different APIs or services

By breaking complex operations into discrete tools and letting the AI agent coordinate between them, you can create flexible, maintainable systems that can evolve with your needs.
