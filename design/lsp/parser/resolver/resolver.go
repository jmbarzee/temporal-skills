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

	if len(errs) > 0 {
		return errs
	}

	// Pass 2: Walk workflow bodies, resolving references.
	for _, def := range file.Definitions {
		wf, ok := def.(*ast.WorkflowDef)
		if !ok {
			continue
		}

		// Build signal and update maps for this workflow.
		signals := make(map[string]*ast.SignalDecl)
		updates := make(map[string]*ast.UpdateDecl)
		for _, s := range wf.Signals {
			signals[s.Name] = s
		}
		for _, u := range wf.Updates {
			updates[u.Name] = u
		}

		ctx := &resolveCtx{
			workflows:  workflows,
			activities: activities,
			signals:    signals,
			updates:    updates,
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

	case *ast.AwaitStmt:
		for _, target := range s.Targets {
			c.resolveAwaitTarget(target)
		}

	case *ast.ParallelBlock:
		c.resolveStatements(s.Body)

	case *ast.SelectBlock:
		for _, sc := range s.Cases {
			c.resolveSelectCase(sc)
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
	}
}

func (c *resolveCtx) resolveAwaitTarget(target *ast.AwaitTarget) {
	switch target.Kind {
	case "signal":
		if def, ok := c.signals[target.Name]; ok {
			target.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined signal: %s", target.Name),
				Line:   target.Line,
				Column: target.Column,
			})
		}
	case "update":
		if def, ok := c.updates[target.Name]; ok {
			target.Resolved = def
		} else {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined update: %s", target.Name),
				Line:   target.Line,
				Column: target.Column,
			})
		}
	}
}

func (c *resolveCtx) resolveSelectCase(sc *ast.SelectCase) {
	switch sc.CaseKind() {
	case "workflow":
		if _, ok := c.workflows[sc.WorkflowName]; !ok {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined workflow: %s", sc.WorkflowName),
				Line:   sc.Line,
				Column: sc.Column,
			})
		}
	case "activity":
		if _, ok := c.activities[sc.ActivityName]; !ok {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined activity: %s", sc.ActivityName),
				Line:   sc.Line,
				Column: sc.Column,
			})
		}
	case "signal":
		if _, ok := c.signals[sc.SignalName]; !ok {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined signal: %s", sc.SignalName),
				Line:   sc.Line,
				Column: sc.Column,
			})
		}
	case "update":
		if _, ok := c.updates[sc.UpdateName]; !ok {
			c.errs = append(c.errs, &ResolveError{
				Msg:    fmt.Sprintf("undefined update: %s", sc.UpdateName),
				Line:   sc.Line,
				Column: sc.Column,
			})
		}
	}
	c.resolveStatements(sc.Body)
}
