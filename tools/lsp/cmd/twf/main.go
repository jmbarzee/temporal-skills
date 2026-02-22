package main

import (
	"fmt"
	"os"
)

const (
	name    = "twf"
	version = "0.1.0"
)

const usage = `twf - Temporal Workflow Format CLI

Usage:
  twf <command> [options] <file...>

Commands:
  check     Parse and validate TWF files
  parse     Output AST as JSON
  symbols   List workflows and activities
  deps      Show dependency graph
  lsp       Start the language server (stdio)
  help      Show this help

Options:
  --lenient  Continue even with resolve errors

Examples:
  twf check workflow.twf
  twf parse workflow.twf
  twf symbols workflow.twf
  twf lsp
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "check":
		os.Exit(checkCommand(os.Args[2:]))
	case "parse":
		os.Exit(parseCommand(os.Args[2:]))
	case "symbols":
		os.Exit(symbolsCommand(os.Args[2:]))
	case "deps":
		os.Exit(depsCommand(os.Args[2:]))
	case "lsp":
		lspCommand()
	case "help", "--help", "-h":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}
