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
		polaris.UseModel("gemini-2.5-pro-preview-05-06"),
		polaris.UseSystemInstruction(
			polaris.AddTextSystemInstruction("You can interact with server logs using available tools."),
		),
		polaris.UseTemperature(0.2),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create AI session: %v", err))
	}
	fmt.Println("AI session created.")

	// Define the prompt for the AI, asking it to use a tool potentially hosted on a remote agent
	prompt := `
        Can you show me the last 5 lines from the ./app.log file?
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
