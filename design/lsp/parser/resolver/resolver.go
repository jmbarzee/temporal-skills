package resolver

import (
	"fmt"

	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
)

// ResolveError represents a resolution error with position info.
type ResolveError struct {
	Msg    string
	Line   int
	Column int
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("resolve error at %d:%d: %s", e.Line, e.Column, e.Msg)
}

// Resolve walks the AST, linking calls to their definitions.
// Returns a list of errors (empty on success).
func Resolve(file *ast.File) []*ResolveError {
	workflows := make(map[string]*ast.WorkflowDef)
	activities := make(map[string]*ast.ActivityDef)
	var errs []*ResolveError

	// Pass 1: Collect all definitions.
	for _, def := range file.Definitions {
		switch d := def.(type) {
		case *ast.WorkflowDef:
			if _, exists := workflows[d.Name]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate workflow definition: %s", d.Name),
					Line:   d.Line,
					Column: d.Column,
				})
			}
			workflows[d.Name] = d
		case *ast.ActivityDef:
			if _, exists := activities[d.Name]; exists {
				errs = append(errs, &ResolveError{
					Msg:    fmt.Sprintf("duplicate activity definition: %s", d.Name),
					Line:   d.Line,
					Column: d.Column,
				})
			}
			activities[d.Name] = d
		}
	}

	// Continue to Pass 2 even if there are duplicate definition errors.
	// This provides better diagnostics by also reporting undefined references.

	// Pass 2: Walk workflow bodies, resolving references.
	for _, def := range file.Definitions {
		wf, ok := def.(*ast.WorkflowDef)
		if !ok {
			continue
		}

		// Build signal, query, and update maps for this workflow.
		signals := make(map[string]*ast.SignalDecl)
		queries := make(map[string]*ast.QueryDecl)
		updates := make(map[string]*ast.UpdateDecl)
		for _, s := range wf.Signals {
			signals[s.Name] = s
		}
		for _, q := range wf.Queries {
			queries[q.Name] = q
		}
		for _, u := range wf.Updates {
			updates[u.Name] = u
		}

		ctx := &resolveCtx{
			workflows:  workflows,
			activities: activities,
			signals:    signals,
			queries:    queries,
			updates:    updates,
		}

		// Resolve handler bodies.
		for _, s := range wf.Signals {
			ctx.resolveStatements(s.Body)
		}
		for _, q := range wf.Queries {
			ctx.resolveStatements(q.Body)
		}
		for _, u := range wf.Updates {
			ctx.resolveStatements(u.Body)
		}

		ctx.resolveStatements(wf.Body)
		errs = append(errs, ctx.errs...)
	}

	return errs
}

type resolveCtx struct {
	workflows  map[string]*ast.WorkflowDef
	activities map[string]*ast.ActivityDef
	signals    map[string]*ast.SignalDecl
	queries    map[string]*ast.QueryDecl
	updates    map[string]*ast.UpdateDecl
	errs       []*ResolveError
}

func (c *resolveCtx) resolveStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		c.resolveStatement(stmt)
	}
}

func (c *resolveCtx) resolveStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ActivityCall:
		if def, ok := c.activities[s.Name]; ok {
			s.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined activity: %s", s.Name),
				Line:   s.Line,
				Column: s.Column,
			})
		}

	case *ast.WorkflowCall:
		if def, ok := c.workflows[s.Name]; ok {
			s.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined workflow: %s", s.Name),
				Line:   s.Line,
				Column: s.Column,
			})
		}

	case *ast.AwaitAllBlock:
		c.resolveStatements(s.Body)

	case *ast.AwaitOneBlock:
		for _, awaitCase := range s.Cases {
			c.resolveAwaitOneCase(awaitCase)
		}

	case *ast.SwitchBlock:
		for _, sc := range s.Cases {
			c.resolveStatements(sc.Body)
		}
		if s.Default != nil {
			c.resolveStatements(s.Default)
		}

	case *ast.IfStmt:
		c.resolveStatements(s.Body)
		if s.ElseBody != nil {
			c.resolveStatements(s.ElseBody)
		}

	case *ast.ForStmt:
		c.resolveStatements(s.Body)

	case *ast.AwaitStmt:
		// Resolve signal/update/activity/workflow references
		if s.Signal != "" {
			if def, ok := c.signals[s.Signal]; ok {
				s.SignalResolved = def
			} else {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined signal: %s", s.Signal),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Update != "" {
			if def, ok := c.updates[s.Update]; ok {
				s.UpdateResolved = def
			} else {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined update: %s", s.Update),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Activity != "" {
			if def, ok := c.activities[s.Activity]; ok {
				s.ActivityResolved = def
			} else {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined activity: %s", s.Activity),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
		if s.Workflow != "" {
			if def, ok := c.workflows[s.Workflow]; ok {
				s.WorkflowResolved = def
			} else {
				c.errs = append(c.errs, &ResolveError{
					Msg:    fmt.Sprintf("undefined workflow: %s", s.Workflow),
					Line:   s.Line,
					Column: s.Column,
				})
			}
		}
	}
}

func (c *resolveCtx) resolveAwaitOneCase(awaitCase *ast.AwaitOneCase) {
	// Resolve signal/update/activity/workflow references
	if awaitCase.Signal != "" {
		if def, ok := c.signals[awaitCase.Signal]; ok {
			awaitCase.SignalResolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined signal: %s", awaitCase.Signal),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}
	if awaitCase.Update != "" {
		if def, ok := c.updates[awaitCase.Update]; ok {
			awaitCase.UpdateResolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined update: %s", awaitCase.Update),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}
	if awaitCase.Activity != "" {
		if def, ok := c.activities[awaitCase.Activity]; ok {
			awaitCase.ActivityResolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined activity: %s", awaitCase.Activity),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}
	if awaitCase.Workflow != "" {
		if def, ok := c.workflows[awaitCase.Workflow]; ok {
			awaitCase.WorkflowResolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined workflow: %s", awaitCase.Workflow),
				Line:   awaitCase.Line,
				Column: awaitCase.Column,
			})
		}
	}

	// Resolve nested await all block if present.
	if awaitCase.AwaitAll != nil {
		c.resolveStatements(awaitCase.AwaitAll.Body)
	}
	// Resolve the case body.
	c.resolveStatements(awaitCase.Body)
}
