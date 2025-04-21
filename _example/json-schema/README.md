# JSON Schema Output in polaris

This example demonstrates two different approaches for generating structured JSON output from AI models in polaris:

1. Using `GenerateJSON` with `UseJSONOutput` for schema-enforced JSON responses
2. Using `Generate` with prompt-based JSON formatting for flexible text outputs

## Overview

When working with AI models, it's often desirable to get structured data rather than free-form text. polaris provides multiple ways to achieve this, each with different trade-offs between strictness and flexibility.

This example showcases both approaches and helps you understand when to use each method based on your specific requirements.

## Prerequisites

Before running this example, you need to set up the following environment variables for Google Cloud Platform authentication:

```bash
export GOOGLE_CLOUD_PROJECT=your_project_id
export GOOGLE_CLOUD_LOCATION=your_gcp_project_location
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credential.json
```

These environment variables are required to authenticate with Google Cloud and access the Gemini AI model used in this example.

## Approach 1: Schema-Enforced JSON with GenerateJSON

The first approach (`outputschema.go`) uses `GenerateJSON` with `UseJSONOutput` to enforce a specific JSON schema:

```go
gen, err := polaris.GenerateJSON(
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
```

### Key Features of Schema-Enforced Approach

- **Type Safety**: The output is guaranteed to match the specified schema
- **Structured Data**: Results are returned as `polaris.Resp` (map[string]any)
- **Validation**: The system automatically validates that the AI's output conforms to the schema
- **Direct Integration**: The structured data can be directly used in your application logic

### Usage

To run this example:

```shell
$ go run outputschema.go
```

The output will be a structured map containing the results of the calculations:

```
map[resultA:56 resultB:4928]
```

## Approach 2: Prompt-Based JSON Formatting

The second approach (`prompt.go`) uses `Generate` with prompt instructions to format the output as JSON:

```go
session, err := polaris.Generate(
    ctx,
    polaris.UseModel("gemini-2.5-pro-exp-03-25"),
    polaris.UseTemperature(0.2),
    // Note: UseJSONOutput is still used here but works differently with Generate
)

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
```

### Key Features of Prompt-Based Approach

- **Flexibility**: The output format can be customized in the prompt
- **Text Output**: Results are returned as text strings
- **Adaptability**: Can handle more complex or dynamic output structures
- **Less Strict**: No automatic validation of the output structure

### Usage

To run this example:

```shell
$ go run prompt.go
```

The output will be a text string containing JSON:

```
{
  "resultA": 56,
  "resultB": 4928
}
```

## Comparing the Two Approaches

### Schema-Enforced Approach (GenerateJSON)

**Advantages:**
- Guarantees type safety and schema conformance
- Returns structured data ready for use in application logic
- Provides automatic validation
- More reliable for critical applications

**Limitations:**
- Less flexible for dynamic or complex output structures
- Requires defining the schema in code

### Prompt-Based Approach (Generate)

**Advantages:**
- More flexible output formatting
- Can adapt to different output requirements without code changes
- Allows for more complex or nested structures
- Output format can be dynamically specified

**Limitations:**
- No automatic type validation
- Results are text strings that need parsing
- Less reliable for ensuring exact output structure
- May require additional error handling

## Conclusion

polaris provides multiple ways to generate structured output from AI models, allowing you to choose the approach that best fits your specific requirements. By understanding the trade-offs between schema enforcement and flexibility, you can make informed decisions about how to structure your AI interactions.

The schema-enforced approach with `GenerateJSON` provides stronger guarantees but less flexibility, while the prompt-based approach with `Generate` offers more flexibility at the cost of strictness. Both approaches have their place in different scenarios, and polaris makes it easy to use either or both as needed.
