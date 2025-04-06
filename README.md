<h1 align="center" style="border-bottom: none">
  <img src="docs/polaris.png" width="150">
</h1>

# polaris: A Distributed AI Agent Framework for Function Calling

[![MIT License](https://img.shields.io/github/license/octu0/polaris)](https://github.com/octu0/polaris/blob/master/LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/octu0/polaris)](https://pkg.go.dev/github.com/octu0/polaris)
[![Go Report Card](https://goreportcard.com/badge/github.com/octu0/polaris)](https://goreportcard.com/report/github.com/octu0/polaris)
[![Releases](https://img.shields.io/github/v/release/octu0/polaris)](https://github.com/octu0/polaris/releases)

`polaris` is a Go framework designed for building **distributed AI agents**.  
These agents act like sidecars, running alongside your applications or directly on servers to expose system capabilities and local resources (like log files or metrics) as secure Function Calls.

This allows AI models (such as Google's Vertex AI Gemini) to intelligently interact with your distributed infrastructure through a unified, simplified Function Calling interface provided by `polaris` agents. The framework is built for parallel execution to handle demanding workloads.

## Features

1.  **Distributed Agent Architecture:** Deploy lightweight `polaris` agents across your infrastructure (servers, containers). Each agent registers specific functions, making local resources or actions available network-wide.
2.  **Access Local Resources via AI:** Enable AI models to securely query log files, fetch system status, execute commands, or interact with other server-local resources through the Function Calls exposed by your distributed agents.
3.  **Parallel Execution:** Handles heavy workloads efficiently by executing incoming Function Call requests in parallel across agents, preventing bottlenecks.
4.  **Simplified JSON Schema:** Define function parameters and responses with a much more concise and readable syntax compared to standard library methods.
5.  **Simple Agent Implementation:** Easily define and register functions ("Tools") within a `polaris` agent to interact with local files, APIs, or system commands.
6.  **Vertex AI Gemini Focused:** Optimized for seamless integration and interaction with Vertex AI **Gemini** models for orchestrating function calls.

## Installation

```bash
go get github.com/octu0/polaris
```

## Simplified JSON Schema Definition

Defining the structure (Schema) for your functions is significantly easier with `polaris` compared to standard Go structures for AI Function Calling.

**Traditional Method (e.g., using `genai` library):**

```go
FunctionDeclarations: []*genai.FunctionDeclaration{
	{
		Name:        "FuncName",
		Description: "FuncDesc",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"param1": {
					Type:        genai.TypeString,
					Description: "desc param1",
				},
				"param2": {
					Type:        genai.TypeInteger,
					Description: "desc param2",
				},
				"param3": {
					Type:        genai.TypeArray,
					Description: "desc param3",
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"param4": {
								Type:        genai.TypeString,
								Description: "desc param4",
							},
							"param5": {
								Type:        genai.TypeBool,
								Description: "desc param5",
							},
						},
					},
				},
			},
			Required: []string{"param1", "param2"},
		},
		Response: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"result1": {
					Type: genai.TypeString,
					Description: "result1",
				},
				"result2": {
					Type: genai.TypeString,
					Description: "result2",
				},
			},
			Required: []string{"result1"},
		},
	},
}
```

**With `polaris`:**


```go
tool := Tool{
	Name:        "FuncName",
	Description: "FuncDesc",
	Parameters: Object{
		Properties: Properties{
			"param1": String{Description: "desc param1", Required: true},
			"param2": Int{Description: "desc param2", Required: true},
			"param3": ObjectArray{
				Description: "desc param3",
				Items: Properties{
					"param4": String{Description: "desc param4"},
					"param5": Bool{Description: "desc param5"},
				},
			},
		},
	},
	Response: Object{
		Properties: Properties{
			"result1": String{Description: "result1", Required: true},
			"result2": String{Description: "result2"},
		},
	},
}
```

## Implementing `polaris` Agent/Tool

You can easily create a standalone agent (or integrate `polaris` into an existing service) that registers specific 'Tools' (Functions). This agent runs, connects to the registry, and listens for requests (orchestrated by the AI) to execute its registered tools.

```go
package main

import (
	"fmt"
	"os"

	"github.com/octu0/polaris"
)

// Example: Tool to read the last N lines of a specific log file
func registerLogReaderAgent(conn *polaris.Conn, logFilePath string) error {
	// Ensure the log file exists (basic check)
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		return fmt.Errorf("log file not found: %s", logFilePath)
	}

	return conn.RegisterTool(polaris.Tool{
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
		// The handler function implements the tool's logic
		Handler: func(ctx *polaris.Ctx) error {
			linesToRead := ctx.Int("lines")
			if linesToRead <= 0 {
				linesToRead = 10 // Use default if invalid
			}

			// --- Placeholder for actual file reading logic ---
			// In a real implementation, you would securely read the
			// last 'linesToRead' lines from 'logFilePath'.
			// Example (conceptual, needs robust implementation):
			// content, err := readLastLines(logFilePath, linesToRead)
			// if err != nil {
			//     return fmt.Errorf("failed to read log file: %w", err)
			// }
			// --- End Placeholder ---

			// Dummy content for example:
			content := fmt.Sprintf("Read last %d lines of %s (implementation pending)", linesToRead, logFilePath)

			ctx.Set(polaris.Resp{
				"log_content": content,
			})
			return nil // Return nil on success
		},
	})
}

func main() {
	// Example: Run an agent exposing a log reader tool
	logFileToMonitor := "/var/log/app.log" // Example log file path

	// Connect the agent to the polaris registry
	conn, err := polaris.Connect(polaris.ConnectAddress("127.0.0.1", "4222"))
	if err != nil {
		panic(fmt.Sprintf("Agent failed to connect: %v", err))
	}
	defer conn.Close()
	fmt.Printf("Agent connected, monitoring %s\n", logFileToMonitor)

	// Register the specific tool this agent provides
	err = registerLogReaderAgent(conn, logFileToMonitor)
	if err != nil {
		panic(fmt.Sprintf("Agent failed to register tool: %v", err))
	}
	fmt.Println("Log reader tool registered successfully.")

	// Keep the agent running to listen for function call requests
	fmt.Println("Agent running...")
	<-make(chan struct{}) // Block forever
}
```

## Usage Example: AI Orchestrating Distributed Agents
From your central application or AI orchestrator service, connect to the `polaris` registry.  
An AI model like Gemini can then discover and invoke functions hosted by your distributed `polaris` agents based on user prompts. The AI doesn't need to know where the agent is running, only that the function is available.

```go
package main

import (
	"context"
	"fmt"

	"github.com/octu0/polaris"
)

func main() {
    ctx := context.Background()

    // Connect to the polaris registry (same network as the agents)
    conn, err := polaris.Connect(polaris.ConnectAddress("127.0.0.1", "4222"))
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    fmt.Println("Orchestrator connected.")

    // Create a session with Vertex AI Gemini
    // Ensure your environment is configured for Vertex AI authentication
    session, err := conn.Use(
        ctx,
        polaris.UseModel("gemini-2.5-pro-exp-03-25"),
        polaris.UseSystemInstruction(
            polaris.AddTextSystemInstruction("You can interact with server logs using available tools."),
        ),
        polaris.UseTemperature(0.2),
    )
    if err != nil {
        panic(fmt.Sprintf("Failed to create AI session: %v", err))
    }
    defer session.Close()
    fmt.Println("AI session created.")

    // Define the prompt for the AI, asking it to use a tool potentially hosted on a remote agent
    prompt := `
        Can you show me the last 5 lines from the /var/log/app.log file?
    `
    fmt.Printf("Sending prompt to AI: %s\n", prompt)

    // Send the prompt. polaris + Gemini will find the 'read_log_file' tool
    // (registered by one of the agents) and attempt to call it.
    it, err := session.SendText(prompt)
    if err != nil {
        panic(fmt.Sprintf("Failed to send prompt: %v", err))
    }

    // Stream and print the AI's response
    fmt.Println("AI Response:")
    for msg, err := range it {
        if err != nil {
            // Handle potential errors during streaming (e.g., function call failure)
            fmt.Printf("Error during response stream: %v\n", err)
            break
        }
        fmt.Println(msg) // Print the content part of the message
    }
    fmt.Println("Interaction complete.")
}
```

## Dependencies

Using `polaris`, AI orchestration capabilities, requires bellow:

1.  **Google Cloud Project:**
    * Access to a Google Cloud project where you can enable APIs and manage resources.
2.  **Vertex AI API Enabled:**
    * The **Vertex AI API** must be enabled within your Google Cloud project.
3.  **Authentication:**
    * **Environment Variables:**
        * `GOOGLE_APPLICATION_CREDENTIALS`: Set this to the path of your service account key JSON file.
        * `GOOGLE_CLOUD_PROJECT`: Set this to your Google Cloud Project ID. 

# License

MIT, see LICENSE file for details.

---

*Portions of this README were generated with the assistance of Google Gemini.*