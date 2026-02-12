package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jmbarzee/temporal-skills/design/lsp/internal/server"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/parser"
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/resolver"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	glspServer "github.com/tliron/glsp/server"
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
  lsp       Start the language server (stdio)
  help      Show this help

Options:
  --json     Output in JSON format (where applicable)
  --lenient  Continue even with resolve errors

Examples:
  twf check workflow.twf
  twf parse --json workflow.twf
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

// parseFiles reads and parses the given files, returning the AST and any errors
// Uses error-tolerant parsing (ParseFileAll) to return partial AST even with parse errors
func parseFiles(args []string) (*ast.File, []string, int) {
	var lenient bool
	var files []string

	for _, arg := range args {
		if arg == "--lenient" {
			lenient = true
		} else if !strings.HasPrefix(arg, "-") {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		return nil, nil, 1
	}

	// Read and concatenate all input files
	var parts []string
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
			return nil, nil, 1
		}
		parts = append(parts, string(data))
	}
	input := strings.Join(parts, "\n")

	// Parse with error tolerance - returns partial AST even with errors
	file, parseErrs := parser.ParseFileAll(input)

	// Collect all error messages
	var allErrs []string

	// Add parse errors
	for _, e := range parseErrs {
		msg := fmt.Sprintf("parse error at %d:%d: %s", e.Line, e.Column, e.Msg)
		allErrs = append(allErrs, msg)
	}

	// Resolve (even if there were parse errors, resolve what we got)
	resolveErrs := resolver.Resolve(file)
	for _, e := range resolveErrs {
		msg := fmt.Sprintf("resolve error at %d:%d: %s", e.Line, e.Column, e.Msg)
		allErrs = append(allErrs, msg)
	}

	// Determine exit code
	exitCode := 0
	if len(allErrs) > 0 && !lenient {
		exitCode = 1
	}

	return file, allErrs, exitCode
}

// checkCommand validates TWF files and reports errors
func checkCommand(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: twf check [--lenient] <file...>")
		return 1
	}

	file, errs, exitCode := parseFiles(args)

	// Always report errors to stderr
	if len(errs) > 0 {
		for _, msg := range errs {
			fmt.Fprintln(os.Stderr, msg)
		}
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

// parseCommand outputs the AST as JSON
// Always outputs partial AST even with errors (lenient by default)
// Errors go to stderr, AST goes to stdout
func parseCommand(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: twf parse [--json] [--lenient] <file...>")
		return 1
	}

	// Force lenient mode for parse command - always emit partial AST
	argsWithLenient := append(args, "--lenient")
	file, errs, _ := parseFiles(argsWithLenient)

	// Output errors to stderr (but don't fail - we still emit JSON)
	for _, msg := range errs {
		fmt.Fprintln(os.Stderr, msg)
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

// symbolsCommand lists all workflows and activities
// Works with partial AST - lists what was successfully parsed
func symbolsCommand(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: twf symbols [--json] [--lenient] <file...>")
		return 1
	}

	var jsonOutput bool
	for _, arg := range args {
		if arg == "--json" {
			jsonOutput = true
			break
		}
	}

	file, errs, exitCode := parseFiles(args)

	// Report errors to stderr but continue to show symbols
	if len(errs) > 0 {
		for _, msg := range errs {
			fmt.Fprintln(os.Stderr, msg)
		}
	}

	// Show symbols from partial AST
	if file != nil {
		if jsonOutput {
			return printSymbolsJSON(file)
		}
		return printSymbolsText(file)
	}

	return exitCode
}

func printSymbolsText(file *ast.File) int {
	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
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
		case *ast.ActivityDef:
			fmt.Printf("activity %s(%s)", d.Name, d.Params)
			if d.ReturnType != "" {
				fmt.Printf(" -> (%s)", d.ReturnType)
			}
			fmt.Println()
		}
	}
	return 0
}

type symbolJSON struct {
	Kind       string   `json:"kind"`
	Name       string   `json:"name"`
	Params     string   `json:"params,omitempty"`
	ReturnType string   `json:"returnType,omitempty"`
	Signals    []string `json:"signals,omitempty"`
	Queries    []string `json:"queries,omitempty"`
	Updates    []string `json:"updates,omitempty"`
}

// lspCommand starts the LSP server over stdio
func lspCommand() {
	commonlog.Configure(1, nil)

	handler, _ := server.NewHandler(name, version)

	s := glspServer.NewServer(handler, name, false)

	s.RunStdio()
}

func printSymbolsJSON(file *ast.File) int {
	var symbols []symbolJSON

	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			sym := symbolJSON{
				Kind:       "workflow",
				Name:       d.Name,
				Params:     d.Params,
				ReturnType: d.ReturnType,
			}
			for _, s := range d.Signals {
				sym.Signals = append(sym.Signals, s.Name)
			}
			for _, q := range d.Queries {
				sym.Queries = append(sym.Queries, q.Name)
			}
			for _, u := range d.Updates {
				sym.Updates = append(sym.Updates, u.Name)
			}
			symbols = append(symbols, sym)
		case *ast.ActivityDef:
			symbols = append(symbols, symbolJSON{
				Kind:       "activity",
				Name:       d.Name,
				Params:     d.Params,
				ReturnType: d.ReturnType,
			})
		}
	}

	data, err := json.MarshalIndent(symbols, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal error: %v\n", err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}
