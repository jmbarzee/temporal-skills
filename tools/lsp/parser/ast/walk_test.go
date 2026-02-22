package ast

import "testing"

func TestWalkStatementsFlat(t *testing.T) {
	stmts := []Statement{
		&RawStmt{Pos: Pos{Line: 1}},
		&ReturnStmt{Pos: Pos{Line: 2}},
		&BreakStmt{Pos: Pos{Line: 3}},
	}

	var lines []int
	WalkStatements(stmts, func(s Statement) bool {
		lines = append(lines, s.NodeLine())
		return true
	})

	if len(lines) != 3 {
		t.Fatalf("expected 3 visits, got %d", len(lines))
	}
	for i, want := range []int{1, 2, 3} {
		if lines[i] != want {
			t.Errorf("visit %d: got line %d, want %d", i, lines[i], want)
		}
	}
}

func TestWalkStatementsNested(t *testing.T) {
	// Build: IfStmt(body=[RawStmt], else=[ForStmt(body=[RawStmt])])
	stmts := []Statement{
		&IfStmt{
			Pos:  Pos{Line: 1},
			Body: []Statement{&RawStmt{Pos: Pos{Line: 2}}},
			ElseBody: []Statement{
				&ForStmt{
					Pos:  Pos{Line: 3},
					Body: []Statement{&RawStmt{Pos: Pos{Line: 4}}},
				},
			},
		},
	}

	var lines []int
	WalkStatements(stmts, func(s Statement) bool {
		lines = append(lines, s.NodeLine())
		return true
	})

	want := []int{1, 2, 3, 4}
	if len(lines) != len(want) {
		t.Fatalf("expected %d visits, got %d: %v", len(want), len(lines), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("visit %d: got line %d, want %d", i, lines[i], want[i])
		}
	}
}

func TestWalkStatementsAwaitOne(t *testing.T) {
	// AwaitOneBlock with two cases; each case has a body statement.
	// The walker should visit: AwaitOneBlock, AwaitOneCase(10), RawStmt(11),
	// AwaitOneCase(20), RawStmt(21).
	stmts := []Statement{
		&AwaitOneBlock{
			Pos: Pos{Line: 1},
			Cases: []*AwaitOneCase{
				{
					Pos:    Pos{Line: 10},
					Target: &TimerTarget{Duration: "1h"},
					Body:   []Statement{&RawStmt{Pos: Pos{Line: 11}}},
				},
				{
					Pos:    Pos{Line: 20},
					Target: &SignalTarget{Signal: Ref[*SignalDecl]{Name: "s"}},
					Body:   []Statement{&RawStmt{Pos: Pos{Line: 21}}},
				},
			},
		},
	}

	var lines []int
	WalkStatements(stmts, func(s Statement) bool {
		lines = append(lines, s.NodeLine())
		return true
	})

	want := []int{1, 10, 11, 20, 21}
	if len(lines) != len(want) {
		t.Fatalf("expected %d visits, got %d: %v", len(want), len(lines), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("visit %d: got line %d, want %d", i, lines[i], want[i])
		}
	}
}

func TestWalkStatementsAwaitOneNestedAwaitAll(t *testing.T) {
	// AwaitOneCase with a nested AwaitAll block.
	stmts := []Statement{
		&AwaitOneBlock{
			Pos: Pos{Line: 1},
			Cases: []*AwaitOneCase{
				{
					Pos: Pos{Line: 10},
					AwaitAll: &AwaitAllBlock{
						Pos:  Pos{Line: 11},
						Body: []Statement{&RawStmt{Pos: Pos{Line: 12}}},
					},
					Body: []Statement{&RawStmt{Pos: Pos{Line: 13}}},
				},
			},
		},
	}

	var lines []int
	WalkStatements(stmts, func(s Statement) bool {
		lines = append(lines, s.NodeLine())
		return true
	})

	want := []int{1, 10, 12, 13}
	if len(lines) != len(want) {
		t.Fatalf("expected %d visits, got %d: %v", len(want), len(lines), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("visit %d: got line %d, want %d", i, lines[i], want[i])
		}
	}
}

func TestWalkStatementsSwitchBlock(t *testing.T) {
	// SwitchBlock with two cases and a default.
	// SwitchCase is a Statement, so the walker visits it before recursing into its body.
	stmts := []Statement{
		&SwitchBlock{
			Pos: Pos{Line: 1},
			Cases: []*SwitchCase{
				{Pos: Pos{Line: 2}, Body: []Statement{&RawStmt{Pos: Pos{Line: 3}}}},
				{Pos: Pos{Line: 4}, Body: []Statement{&RawStmt{Pos: Pos{Line: 5}}}},
			},
			Default: []Statement{&RawStmt{Pos: Pos{Line: 6}}},
		},
	}

	var lines []int
	WalkStatements(stmts, func(s Statement) bool {
		lines = append(lines, s.NodeLine())
		return true
	})

	// Should visit: SwitchBlock(1), SwitchCase(2), RawStmt(3), SwitchCase(4), RawStmt(5), RawStmt(6)
	want := []int{1, 2, 3, 4, 5, 6}
	if len(lines) != len(want) {
		t.Fatalf("expected %d visits, got %d: %v", len(want), len(lines), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("visit %d: got line %d, want %d", i, lines[i], want[i])
		}
	}
}

func TestWalkStatementsEarlyExit(t *testing.T) {
	stmts := []Statement{
		&RawStmt{Pos: Pos{Line: 1}},
		&RawStmt{Pos: Pos{Line: 2}},
		&RawStmt{Pos: Pos{Line: 3}},
	}

	var lines []int
	result := WalkStatements(stmts, func(s Statement) bool {
		lines = append(lines, s.NodeLine())
		return s.NodeLine() != 2 // stop at line 2
	})

	if result != false {
		t.Errorf("expected WalkStatements to return false on early exit")
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 visits before stop, got %d: %v", len(lines), lines)
	}
}

func TestWalkStatementsEarlyExitNested(t *testing.T) {
	stmts := []Statement{
		&IfStmt{
			Pos:  Pos{Line: 1},
			Body: []Statement{&RawStmt{Pos: Pos{Line: 2}}},
		},
		&RawStmt{Pos: Pos{Line: 3}}, // should NOT be visited
	}

	var lines []int
	result := WalkStatements(stmts, func(s Statement) bool {
		lines = append(lines, s.NodeLine())
		return s.NodeLine() != 2
	})

	if result != false {
		t.Errorf("expected WalkStatements to return false on early exit")
	}
	want := []int{1, 2}
	if len(lines) != len(want) {
		t.Fatalf("expected %d visits, got %d: %v", len(want), len(lines), lines)
	}
}

func TestWalkStatementsEmpty(t *testing.T) {
	result := WalkStatements(nil, func(s Statement) bool {
		t.Fatal("should not be called on empty slice")
		return true
	})
	if result != true {
		t.Errorf("expected true for empty walk")
	}
}

func TestAsyncTargetOf(t *testing.T) {
	timer := &TimerTarget{Duration: "5m"}
	signal := &SignalTarget{Signal: Ref[*SignalDecl]{Name: "s"}}
	activity := &ActivityTarget{Activity: Ref[*ActivityDef]{Name: "a"}}

	tests := []struct {
		name string
		stmt Statement
		want AsyncTarget
	}{
		{"AwaitStmt", &AwaitStmt{Target: timer}, timer},
		{"AwaitOneCase", &AwaitOneCase{Target: signal}, signal},
		{"PromiseStmt", &PromiseStmt{Target: activity}, activity},
		{"AwaitOneCase nil target", &AwaitOneCase{}, nil},
		{"RawStmt", &RawStmt{}, nil},
		{"IfStmt", &IfStmt{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AsyncTargetOf(tt.stmt)
			if got != tt.want {
				t.Errorf("AsyncTargetOf(%T) = %v, want %v", tt.stmt, got, tt.want)
			}
		})
	}
}

func TestWalkWithAsyncTargets(t *testing.T) {
	activity := &ActivityTarget{Activity: Ref[*ActivityDef]{Name: "doWork"}}
	stmts := []Statement{
		&AwaitStmt{Pos: Pos{Line: 1}, Target: activity},
		&RawStmt{Pos: Pos{Line: 2}},
	}

	var targets []AsyncTarget
	var parentLines []int
	WalkStatements(stmts, func(s Statement) bool {
		return true
	}, WithAsyncTargets(func(target AsyncTarget, parent Statement) bool {
		targets = append(targets, target)
		parentLines = append(parentLines, parent.NodeLine())
		return true
	}))

	if len(targets) != 1 {
		t.Fatalf("expected 1 async target, got %d", len(targets))
	}
	if targets[0] != activity {
		t.Errorf("expected activity target, got %T", targets[0])
	}
	if parentLines[0] != 1 {
		t.Errorf("expected parent line 1, got %d", parentLines[0])
	}
}

func TestWalkWithAsyncTargetsEarlyExit(t *testing.T) {
	stmts := []Statement{
		&AwaitStmt{Pos: Pos{Line: 1}, Target: &ActivityTarget{Activity: Ref[*ActivityDef]{Name: "a"}}},
		&RawStmt{Pos: Pos{Line: 2}}, // should NOT be visited
	}

	var lines []int
	result := WalkStatements(stmts, func(s Statement) bool {
		lines = append(lines, s.NodeLine())
		return true
	}, WithAsyncTargets(func(target AsyncTarget, parent Statement) bool {
		return false // stop walk
	}))

	if result != false {
		t.Errorf("expected WalkStatements to return false on early exit")
	}
	// fn visits AwaitStmt(1), then async target callback returns false — walk stops
	if len(lines) != 1 {
		t.Fatalf("expected 1 visit before stop, got %d: %v", len(lines), lines)
	}
}

func TestWalkWithAsyncTargetsAwaitOneCase(t *testing.T) {
	signal := &SignalTarget{Signal: Ref[*SignalDecl]{Name: "mySignal"}}
	stmts := []Statement{
		&AwaitOneBlock{
			Pos: Pos{Line: 1},
			Cases: []*AwaitOneCase{
				{
					Pos:    Pos{Line: 10},
					Target: signal,
					Body:   []Statement{&RawStmt{Pos: Pos{Line: 11}}},
				},
			},
		},
	}

	var targets []AsyncTarget
	WalkStatements(stmts, func(s Statement) bool {
		return true
	}, WithAsyncTargets(func(target AsyncTarget, parent Statement) bool {
		targets = append(targets, target)
		return true
	}))

	if len(targets) != 1 {
		t.Fatalf("expected 1 async target, got %d", len(targets))
	}
	if targets[0] != signal {
		t.Errorf("expected signal target, got %T", targets[0])
	}
}

func TestAsyncTargetKind(t *testing.T) {
	tests := []struct {
		target AsyncTarget
		want   string
	}{
		{&TimerTarget{}, "timer"},
		{&SignalTarget{}, "signal"},
		{&UpdateTarget{}, "update"},
		{&ActivityTarget{}, "activity"},
		{&WorkflowTarget{}, "workflow"},
		{&NexusTarget{}, "nexus"},
		{&IdentTarget{}, "ident"},
	}
	for _, tt := range tests {
		got := AsyncTargetKind(tt.target)
		if got != tt.want {
			t.Errorf("AsyncTargetKind(%T) = %q, want %q", tt.target, got, tt.want)
		}
	}
}
