package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
)

type subSymbol struct {
	Name       string `json:"name"`
	Params     string `json:"params,omitempty"`
	ReturnType string `json:"returnType,omitempty"`
}

type symbolJSON struct {
	Kind       string      `json:"kind"`
	Name       string      `json:"name"`
	Params     string      `json:"params,omitempty"`
	ReturnType string      `json:"returnType,omitempty"`
	Signals    []subSymbol `json:"signals,omitempty"`
	Queries    []subSymbol `json:"queries,omitempty"`
	Updates    []subSymbol `json:"updates,omitempty"`
}

// extractSymbols collects workflow and activity definitions into a uniform slice.
func extractSymbols(file *ast.File) []symbolJSON {
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
				sym.Signals = append(sym.Signals, subSymbol{
					Name:   s.Name,
					Params: s.Params,
				})
			}
			for _, q := range d.Queries {
				sym.Queries = append(sym.Queries, subSymbol{
					Name:       q.Name,
					Params:     q.Params,
					ReturnType: q.ReturnType,
				})
			}
			for _, u := range d.Updates {
				sym.Updates = append(sym.Updates, subSymbol{
					Name:       u.Name,
					Params:     u.Params,
					ReturnType: u.ReturnType,
				})
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

	return symbols
}

// symbolsCommand lists all workflows and activities.
// Works with partial AST - lists what was successfully parsed.
func symbolsCommand(args []string) int {
	fs := flag.NewFlagSet("symbols", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	lenient := fs.Bool("lenient", false, "Continue even with resolve errors")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: twf symbols [--json] [--lenient] <file...>")
		return 1
	}

	file, errs, exitCode := parseFiles(paths, *lenient)

	// Report errors to stderr but continue to show symbols
	for _, msg := range errs {
		fmt.Fprintln(os.Stderr, msg)
	}

	// Show symbols from partial AST
	if file != nil {
		if *jsonOutput {
			return printSymbolsJSON(file)
		}
		return printSymbolsText(file)
	}

	return exitCode
}

func printSymbolsText(file *ast.File) int {
	for _, sym := range extractSymbols(file) {
		fmt.Printf("%s %s(%s)", sym.Kind, sym.Name, sym.Params)
		if sym.ReturnType != "" {
			fmt.Printf(" -> (%s)", sym.ReturnType)
		}
		fmt.Println()

		for _, s := range sym.Signals {
			fmt.Printf("  signal %s(%s)\n", s.Name, s.Params)
		}
		for _, q := range sym.Queries {
			fmt.Printf("  query %s(%s)", q.Name, q.Params)
			if q.ReturnType != "" {
				fmt.Printf(" -> (%s)", q.ReturnType)
			}
			fmt.Println()
		}
		for _, u := range sym.Updates {
			fmt.Printf("  update %s(%s)", u.Name, u.Params)
			if u.ReturnType != "" {
				fmt.Printf(" -> (%s)", u.ReturnType)
			}
			fmt.Println()
		}
	}
	return 0
}

func printSymbolsJSON(file *ast.File) int {
	data, err := json.MarshalIndent(extractSymbols(file), "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal error: %v\n", err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}
