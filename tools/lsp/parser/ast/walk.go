package ast

// WalkStatements calls fn on each statement in stmts in pre-order.
// For compound statements, children are visited after the parent.
// If fn returns false, the walk stops immediately.
// Returns false if the walk was stopped early.
func WalkStatements(stmts []Statement, fn func(Statement) bool) bool {
	for _, s := range stmts {
		if !walkStatement(s, fn) {
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
//
// Note: AwaitOneCase implements Statement and is visited directly because each
// case has a distinct AsyncTarget that callers may need to inspect by line.
// SwitchCase does not implement Statement — it is a sub-structure of SwitchBlock,
// and its body statements are visited via SwitchBlock's recursion.
func walkStatement(stmt Statement, fn func(Statement) bool) bool {
	if !fn(stmt) {
		return false
	}
	switch s := stmt.(type) {
	case *AwaitAllBlock:
		if !WalkStatements(s.Body, fn) {
			return false
		}
	case *AwaitOneBlock:
		for _, c := range s.Cases {
			if !fn(c) {
				return false
			}
			if c.AwaitAll != nil {
				if !WalkStatements(c.AwaitAll.Body, fn) {
					return false
				}
			}
			if !WalkStatements(c.Body, fn) {
				return false
			}
		}
	case *SwitchBlock:
		for _, c := range s.Cases {
			if !WalkStatements(c.Body, fn) {
				return false
			}
		}
		if !WalkStatements(s.Default, fn) {
			return false
		}
	case *IfStmt:
		if !WalkStatements(s.Body, fn) {
			return false
		}
		if !WalkStatements(s.ElseBody, fn) {
			return false
		}
	case *ForStmt:
		if !WalkStatements(s.Body, fn) {
			return false
		}
	}
	return true
}
