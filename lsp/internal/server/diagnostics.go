package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func didOpenHandler(store *DocumentStore) protocol.TextDocumentDidOpenFunc {
	return func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		doc := store.Open(params.TextDocument.URI, params.TextDocument.Text)
		return publishDiagnostics(context, doc)
	}
}

func didChangeHandler(store *DocumentStore) protocol.TextDocumentDidChangeFunc {
	return func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		// Full sync: last content change has the full text.
		text := params.ContentChanges[len(params.ContentChanges)-1].(protocol.TextDocumentContentChangeEventWhole).Text
		doc := store.Update(params.TextDocument.URI, text)
		return publishDiagnostics(context, doc)
	}
}

func didCloseHandler(store *DocumentStore) protocol.TextDocumentDidCloseFunc {
	return func(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
		store.Close(params.TextDocument.URI)
		// Clear diagnostics for the closed document.
		context.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: []protocol.Diagnostic{},
		})
		return nil
	}
}

func publishDiagnostics(context *glsp.Context, doc *Document) error {
	var diags []protocol.Diagnostic

	for _, pe := range doc.ParseErrs {
		diags = append(diags, protocol.Diagnostic{
			Range:    posToRange(pe.Line, pe.Column),
			Severity: ptrTo(protocol.DiagnosticSeverityError),
			Source:   ptrTo("twf"),
			Message:  pe.Msg,
		})
	}

	for _, re := range doc.ResolveErrs {
		diags = append(diags, protocol.Diagnostic{
			Range:    posToRange(re.Line, re.Column),
			Severity: ptrTo(protocol.DiagnosticSeverityError),
			Source:   ptrTo("twf"),
			Message:  re.Msg,
		})
	}

	if diags == nil {
		diags = []protocol.Diagnostic{}
	}

	context.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         doc.URI,
		Diagnostics: diags,
	})
	return nil
}

// posToRange converts a 1-based parser position to an LSP 0-based range.
// We highlight the entire line since we don't have end positions.
func posToRange(line, column int) protocol.Range {
	l := uint32(0)
	if line > 0 {
		l = uint32(line - 1)
	}
	c := uint32(0)
	if column > 0 {
		c = uint32(column - 1)
	}
	return protocol.Range{
		Start: protocol.Position{Line: l, Character: c},
		End:   protocol.Position{Line: l, Character: c + 1000}, // highlight to end of line
	}
}
