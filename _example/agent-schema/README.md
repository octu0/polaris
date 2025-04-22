# `agent-schema` Example:

This example demonstrates how to define and use JSON Schema for AI tool outputs in `polaris`. By using `UseJSONOutput` with tool definitions, you can ensure that AI inference outputs conform to your specified schema structure.

## Overview

The `agent-schema` example showcases how to:

- Define tools with structured JSON Schema responses
- Use AI inference to generate outputs that match these schemas
- Create a seamless integration between AI models and structured data

This approach allows you to leverage AI capabilities while maintaining control over the output format, making it easier to process and use the results in your applications.

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

The registry server acts as a central hub for registering and managing tools. It binds to a local address and port, allowing clients and agents to connect.

```go
registry, err := polaris.CreateRegistry(
    polaris.WithBind("127.0.0.1", 4222),
)
```

### 2. Agent Implementation (`agents.go`)

The agent implementation defines several tools with specific JSON schemas for their responses:

- `getWeather`: Returns weather information for a specified city
- `getFortune`: Provides a fortune-telling result
- `getCurrentDate`: Returns the current date information

Each tool uses `polaris.UseJSONOutput(myTool.Response)` to ensure that the AI model's output conforms to the defined schema.

### 3. Client Application (`client.go`)

The client application connects to the registry, creates a session with the AI model, and sends a prompt to get information about today's conditions, including fortune and weather.

## Key Features

### JSON Schema Definition for AI Outputs

The core feature of this example is the ability to define JSON schemas for AI outputs using `UseJSONOutput`. This ensures that the AI model generates responses that match your expected structure.

For example, in the `getWeather` tool:

```go
myTool, _ := conn.Tool(toolName)
gen, err := polaris.GenerateJSON(
    ctx,
    polaris.UseModel("gemini-2.5-pro-exp-03-25"),
    polaris.UseSystemInstruction(
        polaris.AddTextSystemInstruction("Output must be in Japanese."),
    ),
    polaris.UseJSONOutput(myTool.Response),
    polaris.UseTemperature(0.5),
)
```

The `UseJSONOutput(myTool.Response)` parameter tells the AI model to generate output that conforms to the schema defined in `myTool.Response`.

### Structured AI Responses

By using JSON schemas, you can ensure that AI responses are structured and predictable, making them easier to process in your application logic. This approach bridges the gap between free-form AI outputs and structured data requirements.

## Usage

To run this example:

1. Set up the required environment variables as described in the Prerequisites section.

2. Start the registry server:
   ```bash
   go run registry.go
   ```

3. In a separate terminal, start the agent:
   ```bash
   go run agents.go
   ```

4. In another terminal, run the client:
   ```bash
   go run client.go
   ```

The client will send a prompt to the AI model, which will use the registered tools to generate structured responses about today's fortune and weather in Tokyo.

## Code Explanation

### Tool Definition with Response Schema

Each tool is defined with a specific response schema that the AI model must follow:

```go
conn.RegisterTool(polaris.Tool{
    Name:        toolName,
    Description: "get weather by city",
    Parameters: polaris.Object{
        Properties: polaris.Properties{
            "cityName": polaris.String{
                Description: "cityName",
                Default:     "tokyo",
                Required:    true,
            },
        },
    },
    Response: polaris.Object{
        Properties: polaris.Properties{
            "temperature": polaris.Int{
                Description: "estimated maximum temperatures",
                Required:    true,
            },
            "sky_condition": polaris.String{
                Description: "sky condition",
                Required:    true,
            },
        },
    },
    Handler: func(r *polaris.ReqCtx) (polaris.Resp, error) {
        // Handler implementation
    },
})
```

### AI Inference with Schema Enforcement

The `GenerateJSON` function is used to create an AI inference function that enforces the output schema:

```go
gen, err := polaris.GenerateJSON(
    ctx,
    polaris.UseModel("gemini-2.5-pro-exp-03-25"),
    polaris.UseSystemInstruction(
        polaris.AddTextSystemInstruction("Output must be in Japanese."),
    ),
    polaris.UseJSONOutput(myTool.Response),
    polaris.UseTemperature(0.5),
)
```

This ensures that the AI model's output will match the structure defined in `myTool.Response`, allowing for seamless integration between AI capabilities and structured data requirements.

## Benefits

Using JSON schemas for AI outputs provides several benefits:

1. **Predictability**: Ensures that AI outputs follow a consistent structure
2. **Validation**: Automatically validates that outputs contain required fields
3. **Integration**: Makes it easier to integrate AI capabilities into existing systems
4. **Type Safety**: Provides type information for downstream processing

This approach allows you to leverage the power of AI while maintaining the structure and predictability needed for production applications.
