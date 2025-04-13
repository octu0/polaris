package polaris

import (
	"context"
	"fmt"
)

func Example_jsonOutput_promptInstruction() {
	ctx := context.TODO()

	session, err := Generate(
		ctx,
		UseModel("gemini-2.5-pro-exp-03-25"),
		UseTemperature(0.2),
		UseJSONOutput(Object{
			Description: "result of each",
			Properties: Properties{
				"resultA": Int{
					Description: "result 1",
					Required:    true,
				},
				"resultB": Int{
					Description: "result 2",
					Required:    true,
				},
			},
		}),
	)
	if err != nil {
		panic(err)
	}
	defer session.Close()

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
	// Output:
	// {
	//   "resultA": 56,
	//   "resultB": 4928
	// }
	//
}

func Example_jsonOutput_outputSchema() {
	ctx := context.TODO()

	gen, err := GenerateJSON(
		ctx,
		UseModel("gemini-2.5-pro-exp-03-25"),
		UseTemperature(0.2),
		UseJSONOutput(Object{
			Description: "result of each",
			Properties: Properties{
				"resultA": Int{
					Description: "result 1",
					Required:    true,
				},
				"resultB": Int{
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
	`
	resp, err := gen(prompt)
	if err != nil {
		panic(err)
	}
	fmt.Println("resultA=", resp["resultA"])
	fmt.Println("resultB=", resp["resultB"])
	// Output:
	// resultA= 56
	// resultB= 4928
	//
}
