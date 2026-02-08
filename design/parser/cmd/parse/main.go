package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmbarzee/temporal-skills/design/parser/ast"
	"github.com/jmbarzee/temporal-skills/design/parser/parser"
	"github.com/jmbarzee/temporal-skills/design/parser/resolver"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: parse <file.twf> [file2.twf ...]\n")
		os.Exit(1)
	}

	// Read and concatenate all input files.
	var parts []string
	for _, path := range os.Args[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
			os.Exit(1)
		}
		parts = append(parts, string(data))
	}
	input := strings.Join(parts, "\n")

	// Parse.
	file, err := parser.ParseFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	// Resolve.
	errs := resolver.Resolve(file)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "resolve error: %v\n", e)
		}
		os.Exit(1)
	}

	// Print summary.
	var workflows, activities int
	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			workflows++
			fmt.Printf("workflow %s(%s)", d.Name, d.Params)
			if d.ReturnType != "" {
				fmt.Printf(" -> (%s)", d.ReturnType)
			}
			fmt.Println()
			for _, s := range d.Signals {
				fmt.Printf("  signal %s(%s)\n", s.Name, s.Params)
			}
			for _, q := range d.Queries {
				fmt.Printf("  query %s(%s)", q.Name, q.Params)
				if q.ReturnType != "" {
					fmt.Printf(" -> (%s)", q.ReturnType)
				}
				fmt.Println()
			}
			for _, u := range d.Updates {
				fmt.Printf("  update %s(%s)", u.Name, u.Params)
				if u.ReturnType != "" {
					fmt.Printf(" -> (%s)", u.ReturnType)
				}
				fmt.Println()
			}
			fmt.Printf("  %d body statements\n", len(d.Body))
		case *ast.ActivityDef:
			activities++
			fmt.Printf("activity %s(%s)", d.Name, d.Params)
			if d.ReturnType != "" {
				fmt.Printf(" -> (%s)", d.ReturnType)
			}
			fmt.Println()
			fmt.Printf("  %d body statements\n", len(d.Body))
		}
	}
	fmt.Printf("\nOK: %d workflow(s), %d activity(s)\n", workflows, activities)
}
