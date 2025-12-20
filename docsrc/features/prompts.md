# Prompts

Manage version-controlled prompt templates with variable substitution.

## Creating Prompts

```go
// Create a prompt with template
prompt, _ := client.CreatePrompt(ctx, "greeting-prompt",
    opik.WithPromptDescription("Greeting template for users"),
    opik.WithPromptTemplate("Hello, {{name}}! Welcome to {{place}}."),
    opik.WithPromptTags("greeting", "template"),
)
```

## Template Syntax

Use `{{variable}}` syntax for placeholders:

```go
template := `You are a {{role}} assistant.

User query: {{query}}

Please provide a {{style}} response.`
```

## Getting Prompts

```go
// Get latest version by name
version, _ := client.GetPromptByName(ctx, "greeting-prompt", "")

// Get specific version (if you have the commit hash)
version, _ := client.GetPromptByName(ctx, "greeting-prompt", "abc123")
```

## Rendering Templates

```go
// Render with variables
rendered := version.Render(map[string]string{
    "name":  "Alice",
    "place": "Wonderland",
})
// Result: "Hello, Alice! Welcome to Wonderland."

// Extract variables from template
vars := version.ExtractVariables()
// Result: ["name", "place"]
```

## Creating New Versions

```go
// Create a new version of an existing prompt
newVersion, _ := prompt.CreateVersion(ctx, "Hi, {{name}}! Great to see you!",
    opik.WithVersionChangeDescription("Simplified greeting"),
)
```

## Listing Versions

```go
// Get all versions of a prompt
versions, _ := prompt.GetVersions(ctx, 1, 100)

for _, v := range versions {
    fmt.Printf("Version: %s\n", v.Commit)
    fmt.Printf("Template: %s\n", v.Template)
    fmt.Printf("Created: %s\n", v.CreatedAt)
}
```

## Listing All Prompts

```go
prompts, _ := client.ListPrompts(ctx, 1, 100)

for _, p := range prompts {
    fmt.Printf("Prompt: %s\n", p.Name)
    fmt.Printf("Description: %s\n", p.Description)
}
```

## Using Prompts in LLM Calls

```go
func generateResponse(ctx context.Context, client *opik.Client, query string) (string, error) {
    // Get the prompt template
    version, err := client.GetPromptByName(ctx, "assistant-prompt", "")
    if err != nil {
        return "", err
    }

    // Render with variables
    prompt := version.Render(map[string]string{
        "role":  "helpful",
        "query": query,
        "style": "concise",
    })

    // Create trace
    trace, _ := client.Trace(ctx, "generate-response",
        opik.WithTraceInput(map[string]any{
            "query":          query,
            "prompt_name":    "assistant-prompt",
            "prompt_version": version.Commit,
        }),
    )
    defer trace.End(ctx)

    // Call LLM with rendered prompt
    response, err := llmClient.Complete(ctx, prompt)
    if err != nil {
        return "", err
    }

    trace.End(ctx, opik.WithTraceOutput(map[string]any{"response": response}))

    return response, nil
}
```

## Prompt Versioning Best Practices

1. **Use descriptive names**: Make prompts easy to find
2. **Add change descriptions**: Document what changed in each version
3. **Tag prompts**: Use tags for organization (e.g., "production", "experimental")
4. **Track versions in traces**: Include prompt version in trace metadata
5. **Test before promoting**: Use experiments to compare prompt versions
