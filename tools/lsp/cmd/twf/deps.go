package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/deps"
)

// depsCommand extracts and outputs the dependency graph.
func depsCommand(args []string) int {
	fs := flag.NewFlagSet("deps", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	lenient := fs.Bool("lenient", false, "Continue even with resolve errors")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: twf deps [--json] [--lenient] <file...>")
		return 1
	}

	file, errs, exitCode := parseFiles(paths, *lenient)

	printErrors(errs)

	if file == nil {
		return exitCode
	}

	graph := deps.Extract(file)

	if *jsonOutput {
		return printDepsJSON(graph)
	}
	return printDepsText(graph)
}

func printDepsJSON(graph *deps.Graph) int {
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal error: %v\n", err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}

func printDepsText(graph *deps.Graph) int {
	// Summary line.
	s := graph.Summary
	fmt.Printf("Dependency graph: %d workflow(s), %d activity(s), %d service(s), %d edge(s)\n\n",
		s.Workflows, s.Activities, s.NexusServices, s.Edges)

	// Containment.
	if len(graph.Containment) > 0 {
		fmt.Println("Containment:")
		for parent, children := range graph.Containment {
			fmt.Printf("  %s:\n", parent)
			for _, child := range children {
				fmt.Printf("    %s\n", child)
			}
		}
		fmt.Println()
	}

	// Edges grouped by source.
	if len(graph.Edges) > 0 {
		fmt.Println("Edges:")
		grouped := make(map[string][]deps.Edge)
		for _, e := range graph.Edges {
			grouped[e.From] = append(grouped[e.From], e)
		}
		for from, edges := range grouped {
			fmt.Printf("  %s:\n", from)
			for _, e := range edges {
				fmt.Printf("    -> %s (%s, line %d)\n", e.To, e.Kind, e.Line)
			}
		}
		fmt.Println()
	}

	// Cross-worker edges.
	if len(graph.Coarsened.WorkerEdges) > 0 {
		fmt.Println("Cross-worker dependencies:")
		for _, ce := range graph.Coarsened.WorkerEdges {
			fmt.Printf("  %s -> %s (weight: %d)\n", ce.From, ce.To, ce.Weight)
		}
		fmt.Println()
	}

	// Cross-namespace edges.
	if len(graph.Coarsened.NamespaceEdges) > 0 {
		fmt.Println("Cross-namespace dependencies:")
		for _, ce := range graph.Coarsened.NamespaceEdges {
			fmt.Printf("  %s -> %s (weight: %d)\n", ce.From, ce.To, ce.Weight)
		}
		fmt.Println()
	}

	// Unresolved references.
	if len(graph.Unresolved) > 0 {
		fmt.Println("Unresolved references:")
		for _, u := range graph.Unresolved {
			fmt.Printf("  %s -> %s (%s, line %d)\n", u.From, u.Name, u.Kind, u.Line)
		}
		fmt.Println()
	}

	return 0
}
