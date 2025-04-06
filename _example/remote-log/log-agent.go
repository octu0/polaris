package main

import (
	"fmt"
	"os"
	"time"

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
			content := fmt.Sprintf("[%s] Read last %d lines of %s", time.Now(), linesToRead, logFilePath)

			ctx.Set(polaris.Resp{
				"log_content": content,
			})
			return nil
		},
	})
}

func main() {
	// Example: Run an agent exposing a log reader tool
	logFileToMonitor := "./app.log" // Example log file path

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
