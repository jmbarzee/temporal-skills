package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspServer "github.com/tliron/glsp/server"
)

// NewHandler creates a protocol.Handler with all LSP methods registered.
func NewHandler(name, version string) (*protocol.Handler, *DocumentStore) {
	store := NewDocumentStore()

	handler := &protocol.Handler{
		Initialize:  initializeHandler(name, version),
		Initialized: initializedHandler(),
		Shutdown:    shutdownHandler(),
		SetTrace:    setTraceHandler(),

		TextDocumentDidOpen:  didOpenHandler(store),
		TextDocumentDidChange: didChangeHandler(store),
		TextDocumentDidClose: didCloseHandler(store),

		TextDocumentHover:               hoverHandler(store),
		TextDocumentDefinition:          definitionHandler(store),
		TextDocumentDocumentSymbol:      documentSymbolHandler(store),
		TextDocumentCompletion:          completionHandler(store),
		TextDocumentReferences:          referencesHandler(store),
		TextDocumentRename:              renameHandler(store),
		TextDocumentPrepareRename:       prepareRenameHandler(store),
		TextDocumentSemanticTokensFull:  semanticTokensHandler(store),
		TextDocumentFoldingRange:        foldingRangeHandler(store),
		TextDocumentSignatureHelp:       signatureHelpHandler(store),
		TextDocumentCodeAction:          codeActionHandler(store),
	}

	return handler, store
}

// RegisterCustomHandlers registers handlers for LSP features not in protocol_3_16
func RegisterCustomHandlers(s *glspServer.Server, store *DocumentStore) {
	// Register InlayHint handler (LSP 3.17)
	// Note: glsp doesn't expose Handle() publicly, so we'll need to wait for
	// library support or use a different approach
	// TODO: Once glsp supports 3.17, register here:
	// s.Handle("textDocument/inlayHint", inlayHintHandler(store))
}

func initializeHandler(name, version string) protocol.InitializeFunc {
	return func(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
		capabilities := protocol.InitializeResult{
			Capabilities: protocol.ServerCapabilities{
				TextDocumentSync: protocol.TextDocumentSyncOptions{
					OpenClose: boolPtr(true),
					Change:    ptrTo(protocol.TextDocumentSyncKindFull),
				},
				HoverProvider:          &protocol.HoverOptions{},
				DefinitionProvider:     &protocol.DefinitionOptions{},
				DocumentSymbolProvider: &protocol.DocumentSymbolOptions{},
				CompletionProvider:     &protocol.CompletionOptions{},
				ReferencesProvider:     &protocol.ReferenceOptions{},
				RenameProvider:         &protocol.RenameOptions{PrepareProvider: boolPtr(true)},
				FoldingRangeProvider:   &protocol.FoldingRangeOptions{},
				CodeActionProvider: &protocol.CodeActionOptions{
					CodeActionKinds: []protocol.CodeActionKind{
						protocol.CodeActionKindQuickFix,
						protocol.CodeActionKindRefactor,
					},
				},
				SignatureHelpProvider: &protocol.SignatureHelpOptions{
					TriggerCharacters: []string{"("},
				},
				SemanticTokensProvider: &protocol.SemanticTokensOptions{
					Legend: protocol.SemanticTokensLegend{
						TokenTypes:     []string{"keyword", "function", "method", "event", "string", "comment", "operator", "parameter"},
						TokenModifiers: []string{"declaration"},
					},
					Full: true,
				},
			},
			ServerInfo: &protocol.InitializeResultServerInfo{
				Name:    name,
				Version: &version,
			},
		}
		return capabilities, nil
	}
}

func initializedHandler() protocol.InitializedFunc {
	return func(context *glsp.Context, params *protocol.InitializedParams) error {
		return nil
	}
}

func shutdownHandler() protocol.ShutdownFunc {
	return func(context *glsp.Context) error {
		return nil
	}
}

func setTraceHandler() protocol.SetTraceFunc {
	return func(context *glsp.Context, params *protocol.SetTraceParams) error {
		return nil
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func ptrTo[T any](v T) *T {
	return &v
}
