package server

import (
	"strings"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func signatureHelpHandler(store *DocumentStore) protocol.TextDocumentSignatureHelpFunc {
	return func(context *glsp.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
		doc, ok := store.Get(params.TextDocument.URI)
		if !ok || doc.File == nil {
			return nil, nil
		}

		line := int(params.Position.Line) + 1

		node := findNodeAtLine(doc.File, line)
		if node == nil {
			return nil, nil
		}

		switch n := node.(type) {
		case *ast.ActivityCall:
			if n.Activity.Resolved != nil {
				return buildSignatureHelp(n.Activity.Resolved.Name, "activity", n.Activity.Resolved.Params, n.Activity.Resolved.ReturnType), nil
			}
		case *ast.WorkflowCall:
			if n.Workflow.Resolved != nil {
				return buildSignatureHelp(n.Workflow.Resolved.Name, "workflow", n.Workflow.Resolved.Params, n.Workflow.Resolved.ReturnType), nil
			}
		}

		return nil, nil
	}
}

func buildSignatureHelp(name, keyword, params, returnType string) *protocol.SignatureHelp {
	label := keyword + " " + name + "(" + params + ")"
	if returnType != "" {
		label += " -> (" + returnType + ")"
	}

	var parameters []protocol.ParameterInformation
	if params != "" {
		// Offset within the label where params start: after "keyword name("
		paramsOffset := len(keyword) + 1 + len(name) + 1 // "keyword name("
		parts := strings.Split(params, ",")
		offset := paramsOffset
		for _, p := range parts {
			p = strings.TrimSpace(p)
			// Find the actual position of this param in the label
			start := strings.Index(label[offset:], p)
			if start < 0 {
				continue
			}
			start += offset
			end := start + len(p)
			parameters = append(parameters, protocol.ParameterInformation{
				Label: [2]protocol.UInteger{protocol.UInteger(start), protocol.UInteger(end)},
			})
			offset = end
		}
	}

	activeSig := uint32(0)
	return &protocol.SignatureHelp{
		Signatures: []protocol.SignatureInformation{
			{
				Label:      label,
				Parameters: parameters,
			},
		},
		ActiveSignature: &activeSig,
	}
}
