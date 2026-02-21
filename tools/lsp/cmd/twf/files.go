package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/parser"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/resolver"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/validator"
)

// parseFiles reads and parses the given files, returning the AST and any errors.
// Uses error-tolerant parsing (ParseFileAll) to return partial AST even with parse errors.
func parseFiles(paths []string, lenient bool) (*ast.File, []string, int) {
	if len(paths) == 0 {
		return nil, nil, 1
	}

	// Read and concatenate all input files
	var parts []string
	for _, path := range paths {
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
		allErrs = append(allErrs, e.Error())
	}

	// Resolve (even if there were parse errors, resolve what we got)
	resolveErrs := resolver.Resolve(file)
	for _, e := range resolveErrs {
		allErrs = append(allErrs, e.Error())
	}

	// Validate deployment/routing
	validateErrs := validator.Validate(file)
	for _, e := range validateErrs {
		allErrs = append(allErrs, e.Error())
	}

	// Determine exit code
	exitCode := 0
	if len(allErrs) > 0 && !lenient {
		exitCode = 1
	}

	return file, allErrs, exitCode
}
