package server

import (
	"github.com/jmbarzee/temporal-skills/design/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func completionHandler(store *DocumentStore) protocol.TextDocumentCompletionFunc {
	return func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil {
			return nil, nil
		}

		line := int(params.Position.Line) + 1 // LSP 0-based → parser 1-based

		ctx := findCompletionContext(doc.File, line)

		var items []protocol.CompletionItem
		switch ctx.kind {
		case contextTopLevel:
			items = topLevelCompletions()
		case contextWorkflow:
			items = workflowCompletions(doc.File, ctx.workflow)
		case contextActivity:
			items = activityCompletions()
		}

		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        items,
		}, nil
	}
}

type completionContextKind int

const (
	contextTopLevel completionContextKind = iota
	contextWorkflow
	contextActivity
)

type completionContext struct {
	kind     completionContextKind
	workflow *ast.WorkflowDef // non-nil when kind == contextWorkflow
}

// findCompletionContext determines whether the cursor line falls inside a
// workflow body, activity body, or at the top level. It uses the line ranges
// of definitions from the AST to decide.
func findCompletionContext(file *ast.File, line int) completionContext {
	if file == nil {
		return completionContext{kind: contextTopLevel}
	}

	for i, def := range file.Definitions {
		startLine := def.NodeLine()

		// Determine the end boundary: the line before the next definition starts,
		// or a very large number for the last definition.
		endLine := 1<<31 - 1
		if i+1 < len(file.Definitions) {
			endLine = file.Definitions[i+1].NodeLine() - 1
		}

		if line > startLine && line <= endLine {
			switch d := def.(type) {
			case *ast.WorkflowDef:
				return completionContext{kind: contextWorkflow, workflow: d}
			case *ast.ActivityDef:
				return completionContext{kind: contextActivity}
			}
		}
	}

	return completionContext{kind: contextTopLevel}
}

func topLevelCompletions() []protocol.CompletionItem {
	return []protocol.CompletionItem{
		keywordItem("workflow", "Define a new workflow"),
		keywordItem("activity", "Define a new activity"),
	}
}

func workflowCompletions(file *ast.File, enclosing *ast.WorkflowDef) []protocol.CompletionItem {
	items := []protocol.CompletionItem{
		keywordItem("activity", "Call an activity"),
		keywordItem("workflow", "Call a child workflow"),
		keywordItem("spawn", "Spawn a workflow asynchronously"),
		keywordItem("detach", "Detach a fire-and-forget workflow"),
		keywordItem("nexus", "Call a workflow via Nexus"),
		keywordItem("timer", "Wait for a duration"),
		keywordItem("await", "Await a signal or update"),
		keywordItem("parallel", "Run statements in parallel"),
		keywordItem("select", "Race between branches"),
		keywordItem("switch", "Switch on an expression"),
		keywordItem("if", "Conditional statement"),
		keywordItem("for", "Loop statement"),
		keywordItem("return", "Return a value"),
		keywordItem("continue_as_new", "Continue as new execution"),
		keywordItem("break", "Break out of a loop"),
		keywordItem("continue", "Continue to next iteration"),
		keywordItem("signal", "Declare a signal handler"),
		keywordItem("query", "Declare a query handler"),
		keywordItem("update", "Declare an update handler"),
	}

	// Add defined activity/workflow names as completion targets.
	if file != nil {
		for _, def := range file.Definitions {
			switch d := def.(type) {
			case *ast.ActivityDef:
				items = append(items, nameItem(d.Name, "Activity definition"))
			case *ast.WorkflowDef:
				if enclosing == nil || d.Name != enclosing.Name {
					items = append(items, nameItem(d.Name, "Workflow definition"))
				}
			}
		}
	}

	// Add signal/update names from the enclosing workflow.
	if enclosing != nil {
		for _, s := range enclosing.Signals {
			items = append(items, nameItem(s.Name, "Signal"))
		}
		for _, u := range enclosing.Updates {
			items = append(items, nameItem(u.Name, "Update"))
		}
	}

	return items
}

func activityCompletions() []protocol.CompletionItem {
	return []protocol.CompletionItem{
		keywordItem("switch", "Switch on an expression"),
		keywordItem("if", "Conditional statement"),
		keywordItem("for", "Loop statement"),
		keywordItem("return", "Return a value"),
		keywordItem("break", "Break out of a loop"),
		keywordItem("continue", "Continue to next iteration"),
	}
}

func keywordItem(kw, detail string) protocol.CompletionItem {
	kind := protocol.CompletionItemKindKeyword
	return protocol.CompletionItem{
		Label:  kw,
		Kind:   &kind,
		Detail: &detail,
	}
}

func nameItem(name, detail string) protocol.CompletionItem {
	kind := protocol.CompletionItemKindReference
	return protocol.CompletionItem{
		Label:  name,
		Kind:   &kind,
		Detail: &detail,
	}
}
