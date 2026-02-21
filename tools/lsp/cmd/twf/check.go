package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
)

// checkCommand validates TWF files and reports errors.
func checkCommand(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	lenient := fs.Bool("lenient", false, "Continue even with resolve errors")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: twf check [--lenient] <file...>")
		return 1
	}

	file, errs, exitCode := parseFiles(paths, *lenient)

	// Always report errors to stderr
	for _, msg := range errs {
		fmt.Fprintln(os.Stderr, msg)
	}

	// Count definitions from partial AST
	var workflows, activities int
	if file != nil {
		for _, def := range file.Definitions {
			switch def.(type) {
			case *ast.WorkflowDef:
				workflows++
			case *ast.ActivityDef:
				activities++
			}
		}
	}

	if exitCode != 0 {
		// Still show what we parsed
		if workflows > 0 || activities > 0 {
			fmt.Fprintf(os.Stderr, "Partial parse: %d workflow(s), %d activity(s)\n", workflows, activities)
		}
		return exitCode
	}

	fmt.Printf("✓ OK: %d workflow(s), %d activity(s)\n", workflows, activities)
	return 0
}
