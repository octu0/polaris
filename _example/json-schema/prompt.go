package main

import (
	"context"
	"fmt"

	"github.com/octu0/polaris"
)

func main() {
	ctx := context.TODO()

	session, err := polaris.Generate(
		ctx,
		polaris.UseModel("gemini-2.5-pro-exp-03-25"),
		polaris.UseTemperature(0.2),
		polaris.UseJSONOutput(polaris.Object{
			Description: "result of each",
			Properties: polaris.Properties{
				"resultA": polaris.Int{
					Description: "result 1",
					Required:    true,
				},
				"resultB": polaris.Int{
					Description: "result 2",
					Required:    true,
				},
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	prompt := `
		execute this task:
		1. Sum 35 and 21
		2. multiply by 88 using one previous answer.

		output(JSON schema):
		ret={
		  "resultA": int,
		  "resultB": int,
		}
		Return: ret
	`
	it, err := session.SendText(prompt)
	if err != nil {
		panic(err)
	}
	for msg, err := range it {
		if err != nil {
			panic(err)
		}
		fmt.Println(msg)
	}
	// Outputs:
	// {
	//   "resultA": 56,
	//   "resultB": 4928
	// }
	//
}
