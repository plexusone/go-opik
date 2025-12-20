package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	opik "github.com/grokify/go-comet-ml-opik"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "configure":
		runConfigure(args)
	case "projects":
		runProjects(args)
	case "traces":
		runTraces(args)
	case "datasets":
		runDatasets(args)
	case "experiments":
		runExperiments(args)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Opik CLI - LLM Observability Tool

Usage:
  opik <command> [options]

Commands:
  configure    Configure Opik credentials
  projects     Manage projects
  traces       View and manage traces
  datasets     Manage datasets
  experiments  Manage experiments
  help         Show this help message

Use "opik <command> -h" for more information about a command.

Environment Variables:
  OPIK_API_KEY      API key for Opik Cloud
  OPIK_WORKSPACE    Workspace name
  OPIK_URL_OVERRIDE Custom API endpoint URL
  OPIK_PROJECT_NAME Default project name`)
}

func runConfigure(args []string) {
	fs := flag.NewFlagSet("configure", flag.ExitOnError)
	apiKey := fs.String("api-key", "", "API key for Opik Cloud")
	workspace := fs.String("workspace", "", "Workspace name")
	url := fs.String("url", "", "Custom API endpoint URL")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Check current configuration
	cfg := opik.LoadConfig()

	if *apiKey != "" {
		cfg.APIKey = *apiKey
	}
	if *workspace != "" {
		cfg.Workspace = *workspace
	}
	if *url != "" {
		cfg.URL = *url
	}

	// Save to config file
	err := opik.SaveConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully.")
	fmt.Printf("  URL: %s\n", cfg.URL)
	fmt.Printf("  Workspace: %s\n", cfg.Workspace)
	if cfg.APIKey != "" {
		fmt.Printf("  API Key: %s...\n", cfg.APIKey[:min(8, len(cfg.APIKey))])
	}
}

func runProjects(args []string) {
	fs := flag.NewFlagSet("projects", flag.ExitOnError)
	list := fs.Bool("list", false, "List all projects")
	create := fs.String("create", "", "Create a new project with the given name")
	format := fs.String("format", "text", "Output format (text, json)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := opik.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	if *list {
		projects, err := client.ListProjects(ctx, 1, 100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
			os.Exit(1)
		}

		if *format == "json" {
			_ = json.NewEncoder(os.Stdout).Encode(projects)
		} else {
			fmt.Println("Projects:")
			for _, p := range projects {
				fmt.Printf("  - %s (ID: %s)\n", p.Name, p.ID)
			}
		}
		return
	}

	if *create != "" {
		project, err := client.CreateProject(ctx, *create)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created project: %s (ID: %s)\n", project.Name, project.ID)
		return
	}

	fs.Usage()
}

func runTraces(args []string) {
	fs := flag.NewFlagSet("traces", flag.ExitOnError)
	list := fs.Bool("list", false, "List recent traces")
	project := fs.String("project", "", "Filter by project name")
	limit := fs.Int("limit", 10, "Maximum number of traces to show")
	format := fs.String("format", "text", "Output format (text, json)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	opts := []opik.Option{}
	if *project != "" {
		opts = append(opts, opik.WithProjectName(*project))
	}

	client, err := opik.NewClient(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	if *list {
		traces, err := client.ListTraces(ctx, 1, *limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing traces: %v\n", err)
			os.Exit(1)
		}

		if *format == "json" {
			_ = json.NewEncoder(os.Stdout).Encode(traces)
		} else {
			fmt.Printf("Recent traces (limit %d):\n", *limit)
			for _, t := range traces {
				duration := ""
				if !t.EndTime.IsZero() {
					duration = fmt.Sprintf(" (%.2fs)", t.EndTime.Sub(t.StartTime).Seconds())
				}
				fmt.Printf("  - %s: %s%s\n", t.ID[:8], t.Name, duration)
			}
		}
		return
	}

	fs.Usage()
}

func runDatasets(args []string) {
	fs := flag.NewFlagSet("datasets", flag.ExitOnError)
	list := fs.Bool("list", false, "List all datasets")
	create := fs.String("create", "", "Create a new dataset with the given name")
	get := fs.String("get", "", "Get a dataset by name")
	deleteFlag := fs.String("delete", "", "Delete a dataset by name")
	format := fs.String("format", "text", "Output format (text, json)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := opik.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	if *list {
		datasets, err := client.ListDatasets(ctx, 1, 100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing datasets: %v\n", err)
			os.Exit(1)
		}

		if *format == "json" {
			_ = json.NewEncoder(os.Stdout).Encode(datasets)
		} else {
			fmt.Println("Datasets:")
			for _, d := range datasets {
				fmt.Printf("  - %s (ID: %s)\n", d.Name(), d.ID())
			}
		}
		return
	}

	if *create != "" {
		dataset, err := client.CreateDataset(ctx, *create)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating dataset: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created dataset: %s (ID: %s)\n", dataset.Name(), dataset.ID())
		return
	}

	if *get != "" {
		dataset, err := client.GetDatasetByName(ctx, *get)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting dataset: %v\n", err)
			os.Exit(1)
		}

		if *format == "json" {
			_ = json.NewEncoder(os.Stdout).Encode(dataset)
		} else {
			fmt.Printf("Dataset: %s\n", dataset.Name())
			fmt.Printf("  ID: %s\n", dataset.ID())
			fmt.Printf("  Description: %s\n", dataset.Description())
		}
		return
	}

	if *deleteFlag != "" {
		dataset, err := client.GetDatasetByName(ctx, *deleteFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding dataset: %v\n", err)
			os.Exit(1)
		}
		err = dataset.Delete(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting dataset: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted dataset: %s\n", *deleteFlag)
		return
	}

	fs.Usage()
}

func runExperiments(args []string) {
	fs := flag.NewFlagSet("experiments", flag.ExitOnError)
	list := fs.Bool("list", false, "List experiments")
	dataset := fs.String("dataset", "", "Filter by dataset name")
	format := fs.String("format", "text", "Output format (text, json)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := opik.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	if *list {
		if *dataset == "" {
			fmt.Fprintf(os.Stderr, "Error: -dataset is required for listing experiments\n")
			os.Exit(1)
		}

		ds, err := client.GetDatasetByName(ctx, *dataset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding dataset: %v\n", err)
			os.Exit(1)
		}

		experiments, err := client.ListExperiments(ctx, ds.ID(), 1, 100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing experiments: %v\n", err)
			os.Exit(1)
		}

		if *format == "json" {
			_ = json.NewEncoder(os.Stdout).Encode(experiments)
		} else {
			fmt.Printf("Experiments for dataset '%s':\n", *dataset)
			for _, e := range experiments {
				fmt.Printf("  - %s (ID: %s)\n", e.Name(), e.ID())
			}
		}
		return
	}

	fs.Usage()
}
