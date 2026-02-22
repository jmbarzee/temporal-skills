package ast

// WalkOption configures optional behavior for WalkStatements.
type WalkOption func(*walkConfig)

type walkConfig struct {
	asyncTargetFn func(AsyncTarget, Statement) bool
}

// WithAsyncTargets registers a callback that is invoked for each AsyncTarget
// found in AwaitStmt, AwaitOneCase, and PromiseStmt nodes. The callback
// receives both the target and its parent statement. If the callback returns
// false, the walk stops immediately.
func WithAsyncTargets(fn func(target AsyncTarget, parent Statement) bool) WalkOption {
	return func(cfg *walkConfig) { cfg.asyncTargetFn = fn }
}

// WalkStatements calls fn on each statement in stmts in pre-order.
// For compound statements, children are visited after the parent.
// If fn returns false, the walk stops immediately.
// Returns false if the walk was stopped early.
func WalkStatements(stmts []Statement, fn func(Statement) bool, opts ...WalkOption) bool {
	var cfg walkConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	for _, s := range stmts {
		if !walkStatement(s, fn, &cfg) {
			return false
		}
	}
	return true
}

// AsyncTargetOf returns the AsyncTarget embedded in a statement, or nil.
// Statements that contain an AsyncTarget: AwaitStmt, AwaitOneCase, PromiseStmt.
func AsyncTargetOf(s Statement) AsyncTarget {
	switch n := s.(type) {
	case *AwaitStmt:
		return n.Target
	case *AwaitOneCase:
		return n.Target
	case *PromiseStmt:
		return n.Target
	}
	return nil
}

// walkStatement visits a single statement and recursively visits its children.
func walkStatement(stmt Statement, fn func(Statement) bool, cfg *walkConfig) bool {
	if !fn(stmt) {
		return false
	}

	// Invoke async target callback if configured.
	if cfg.asyncTargetFn != nil {
		if target := AsyncTargetOf(stmt); target != nil {
			if !cfg.asyncTargetFn(target, stmt) {
				return false
			}
		}
	}

	switch s := stmt.(type) {
	case *AwaitAllBlock:
		for _, child := range s.Body {
			if !walkStatement(child, fn, cfg) {
				return false
			}
		}
	case *AwaitOneBlock:
		for _, c := range s.Cases {
			if !walkStatement(c, fn, cfg) {
				return false
			}
		}
	case *AwaitOneCase:
		if s.AwaitAll != nil {
			for _, child := range s.AwaitAll.Body {
				if !walkStatement(child, fn, cfg) {
					return false
				}
			}
		}
		for _, child := range s.Body {
			if !walkStatement(child, fn, cfg) {
				return false
			}
		}
	case *SwitchBlock:
		for _, c := range s.Cases {
			if !walkStatement(c, fn, cfg) {
				return false
			}
		}
		for _, child := range s.Default {
			if !walkStatement(child, fn, cfg) {
				return false
			}
		}
	case *SwitchCase:
		for _, child := range s.Body {
			if !walkStatement(child, fn, cfg) {
				return false
			}
		}
	case *IfStmt:
		for _, child := range s.Body {
			if !walkStatement(child, fn, cfg) {
				return false
			}
		}
		for _, child := range s.ElseBody {
			if !walkStatement(child, fn, cfg) {
				return false
			}
		}
	case *ForStmt:
		for _, child := range s.Body {
			if !walkStatement(child, fn, cfg) {
				return false
			}
		}
	}
	return true
}
