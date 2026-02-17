package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func renameHandler(store *DocumentStore) protocol.TextDocumentRenameFunc {
	return func(context *glsp.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		line := int(params.Position.Line) + 1

		node := findNodeAtLine(doc.File, line)
		if node == nil {
			return nil, nil
		}

		name, kind := nameOfNode(node)
		if name == "" {
			return nil, nil
		}

		refs := collectReferences(doc.File, name, kind, true)
		if len(refs) == 0 {
			return nil, nil
		}

		var edits []protocol.TextEdit
		for _, ref := range refs {
			edits = append(edits, protocol.TextEdit{
				Range:   nameRange(ref),
				NewText: params.NewName,
			})
		}

		return &protocol.WorkspaceEdit{
			Changes: map[string][]protocol.TextEdit{
				params.TextDocument.URI: edits,
			},
		}, nil
	}
}

func prepareRenameHandler(store *DocumentStore) protocol.TextDocumentPrepareRenameFunc {
	return func(context *glsp.Context, params *protocol.PrepareRenameParams) (any, error) {
		doc := store.Get(params.TextDocument.URI)
		if doc == nil || doc.File == nil {
			return nil, nil
		}

		line := int(params.Position.Line) + 1

		node := findNodeAtLine(doc.File, line)
		if node == nil {
			return nil, nil
		}

		name, kind := nameOfNode(node)
		if name == "" || kind == "" {
			return nil, nil
		}

		return nameRange(node), nil
	}
}

