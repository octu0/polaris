# Remote Log Access with polaris

This example demonstrates a basic implementation of using polaris to access remote log files through natural language queries. It showcases how an AI agent can interact with a specialized tool that provides access to server logs.

## Overview

In many real-world scenarios, accessing and analyzing log files is a common task for developers and system administrators. This example shows how polaris can be used to:

1. Create a specialized agent that provides access to log files
2. Allow users to query log content using natural language
3. Bridge the gap between human language and system operations

This represents one of the most basic and practical use cases for polaris: enabling natural language interfaces to system operations.

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

The registry server acts as a central hub for registering and managing tools. It binds to a local address and port, allowing clients and agents to connect.

```go
registry, err := polaris.CreateRegistry(
    polaris.WithBind("127.0.0.1", 4222),
)
```

### 2. Log Agent (`log-agent.go`)

The log agent registers a tool that can read the last N lines from a specified log file:

```go
conn.RegisterTool(polaris.Tool{
    Name:        "read_log_file",
    Description: fmt.Sprintf("Reads the last N lines from the log file: %s", logFilePath),
    Parameters: polaris.Object{
        Properties: polaris.Properties{
            "lines": polaris.Int{
                Description: "Number of lines to read from the end of the file",
                Required:    true,
                Default:     10,
            },
        },
    },
    Response: polaris.Object{
        Properties: polaris.Properties{
            "log_content": polaris.String{
                Description: "The last N lines of the log file",
                Required:    true,
            },
        },
    },
    Handler: func(r *polaris.ReqCtx) (polaris.Resp, error) {
        linesToRead := r.Int("lines")
        // Implementation to read log file...
        return polaris.Resp{
            "log_content": content,
        }, nil
    },
})
```

### 3. Client Application (`client.go`)

The client application connects to the registry, creates a session with the AI model, and sends a natural language prompt to request log information:

```go
prompt := `
    Can you show me the last 5 lines from the ./app.log file?
`
```

### 4. Sample Log File (`app.log`)

A simple log file that serves as the target for the log reading operations. In this example, it contains minimal content for demonstration purposes.

## How Remote Log Access Works

When the user sends a natural language query about logs:

1. The AI model processes the query and understands the intent to read log files
2. It identifies the appropriate tool (`read_log_file`) to fulfill this request
3. It extracts parameters from the query (e.g., the number of lines to read)
4. It calls the tool with the appropriate parameters
5. The log agent executes the request and returns the log content
6. The AI model formats the response in a human-readable way

This process demonstrates how polaris bridges the gap between natural language and system operations, making it easier for users to access information without needing to remember specific commands or syntax.

## Usage

To run this example:

1. Set up the required environment variables as described in the Prerequisites section.

2. Start the registry server:

```shell
$ go run registry.go
```

3. In a separate terminal, start the log agent:

```shell
$ go run log-agent.go
```

4. In another terminal, run the client:

```shell
$ go run client.go
```

## Expected Output

When you run this example, the client will output a response that includes:

1. Confirmation of connection to the registry
2. Creation of the AI session
3. The prompt being sent
4. The AI's response, which will include the content of the log file (or in this example, a placeholder message indicating that the last N lines were read)

The exact response will depend on the AI model's generation, but it will typically include:
- Acknowledgment of the request
- The log content (or a message about it)
- Possibly some context or explanation about the log content

## Real-World Applications

This pattern of using natural language to access system information can be extended to many practical applications:

- Monitoring system health and performance
- Troubleshooting application issues
- Analyzing security logs for potential threats
- Generating reports based on log data
- Automating routine system administration tasks

## Benefits of Natural Language Log Access

Using Polaris to provide natural language access to logs offers several advantages:

1. **Accessibility**: Users don't need to remember specific log file locations or command syntax
2. **Efficiency**: Complex log queries can be expressed in simple natural language
3. **Context-Awareness**: The AI can understand the intent behind queries and provide relevant information
4. **Extensibility**: The same pattern can be applied to other system operations beyond log access

This example demonstrates just one of the many ways Polaris can be used to create more intuitive interfaces to technical systems and operations.
