package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// parseCommand outputs the AST as JSON.
// Always outputs partial AST even with errors (lenient by default).
// Errors go to stderr, AST goes to stdout.
func parseCommand(args []string) int {
	fs := flag.NewFlagSet("parse", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: twf parse <file...>")
		return 1
	}

	// Force lenient mode - always emit partial AST
	file, errs, _ := parseFiles(paths, true)

	// Output errors to stderr (but don't fail - we still emit JSON)
	printErrors(errs)

	if file == nil {
		fmt.Println("null")
		return 1
	}

	// Output AST to stdout even if there were errors
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal error: %v\n", err)
		return 1
	}
	fmt.Println(string(data))

	// Exit 0 even with parse/resolve errors - the visualizer needs the partial AST
	return 0
}
