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
