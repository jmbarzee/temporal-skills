package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/parser"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/resolver"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/validator"
)

// parseFiles reads and parses the given files, returning the AST and any errors.
// Each file is parsed independently with per-file line numbers. Definitions are
// stamped with their source file and merged into a single AST for resolution.
func parseFiles(paths []string, lenient bool) (*ast.File, []string, int) {
	if len(paths) == 0 {
		return nil, nil, 1
	}

	merged := &ast.File{}
	var allErrs []string

	// Parse each file independently
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
			return nil, nil, 1
		}

		file, parseErrs := parser.ParseFileAll(string(data))

		// Collect parse errors with filename prefix
		for _, e := range parseErrs {
			allErrs = append(allErrs, fmt.Sprintf("%s: %s", filepath.Base(path), e.Error()))
		}

		// Stamp source file and merge definitions
		base := filepath.Base(path)
		for _, def := range file.Definitions {
			setSourceFile(def, base)
			merged.Definitions = append(merged.Definitions, def)
		}
	}

	// Resolve across all files
	resolveErrs := resolver.Resolve(merged)
	for _, e := range resolveErrs {
		allErrs = append(allErrs, e.Error())
	}

	// Validate deployment/routing
	validateErrs := validator.Validate(merged)
	for _, e := range validateErrs {
		allErrs = append(allErrs, e.Error())
	}

	// Determine exit code
	exitCode := 0
	if len(allErrs) > 0 && !lenient {
		exitCode = 1
	}

	return merged, allErrs, exitCode
}

// setSourceFile stamps a definition with its source file name.
func setSourceFile(def ast.Definition, sourceFile string) {
	switch d := def.(type) {
	case *ast.WorkflowDef:
		d.SourceFile = sourceFile
	case *ast.ActivityDef:
		d.SourceFile = sourceFile
	case *ast.WorkerDef:
		d.SourceFile = sourceFile
	case *ast.NamespaceDef:
		d.SourceFile = sourceFile
	case *ast.NexusServiceDef:
		d.SourceFile = sourceFile
	}
}

// printErrors writes error messages to stderr.
func printErrors(errs []string) {
	for _, msg := range errs {
		fmt.Fprintln(os.Stderr, msg)
	}
}
