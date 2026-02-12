package server

import (
	"github.com/tliron/glsp"
	protocol316 "github.com/tliron/glsp/protocol_3_16"
	protocol "github.com/tliron/glsp/protocol_3_17"
)

// NewHandler creates a protocol.Handler with all LSP methods registered.
func NewHandler(name, version string) (*protocol.Handler, *DocumentStore) {
	store := NewDocumentStore()

	handler := &protocol.Handler{
		Handler: protocol316.Handler{
			Initialized: initializedHandler(),
			Shutdown:    shutdownHandler(),
			SetTrace:    setTraceHandler(),

			TextDocumentDidOpen:  didOpenHandler(store),
			TextDocumentDidChange: didChangeHandler(store),
			TextDocumentDidClose:  didCloseHandler(store),

			TextDocumentHover:              hoverHandler(store),
			TextDocumentDefinition:         definitionHandler(store),
			TextDocumentDocumentSymbol:     documentSymbolHandler(store),
			TextDocumentCompletion:         completionHandler(store),
			TextDocumentReferences:         referencesHandler(store),
			TextDocumentRename:             renameHandler(store),
			TextDocumentPrepareRename:      prepareRenameHandler(store),
			TextDocumentSemanticTokensFull: semanticTokensHandler(store),
			TextDocumentFoldingRange:       foldingRangeHandler(store),
			TextDocumentSignatureHelp:      signatureHelpHandler(store),
			TextDocumentCodeAction:         codeActionHandler(store),
		},
		Initialize:            initializeHandler(name, version),
		TextDocumentInlayHint: inlayHintHandler(store),
	}

	return handler, store
}

func initializeHandler(name, version string) protocol.InitializeFunc {
	return func(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
		capabilities := protocol.InitializeResult{
			Capabilities: protocol.ServerCapabilities{
				ServerCapabilities: protocol316.ServerCapabilities{
					TextDocumentSync: protocol316.TextDocumentSyncOptions{
						OpenClose: boolPtr(true),
						Change:    ptrTo(protocol316.TextDocumentSyncKindFull),
					},
					HoverProvider:          &protocol316.HoverOptions{},
					DefinitionProvider:     &protocol316.DefinitionOptions{},
					DocumentSymbolProvider: &protocol316.DocumentSymbolOptions{},
					CompletionProvider:     &protocol316.CompletionOptions{},
					ReferencesProvider:     &protocol316.ReferenceOptions{},
					RenameProvider:         &protocol316.RenameOptions{PrepareProvider: boolPtr(true)},
					FoldingRangeProvider:   &protocol316.FoldingRangeOptions{},
					CodeActionProvider: &protocol316.CodeActionOptions{
						CodeActionKinds: []protocol316.CodeActionKind{
							protocol316.CodeActionKindQuickFix,
							protocol316.CodeActionKindRefactor,
						},
					},
					SignatureHelpProvider: &protocol316.SignatureHelpOptions{
						TriggerCharacters: []string{"("},
					},
					SemanticTokensProvider: &protocol316.SemanticTokensOptions{
						Legend: protocol316.SemanticTokensLegend{
							TokenTypes:     []string{"keyword", "function", "method", "event", "string", "comment", "operator", "parameter"},
							TokenModifiers: []string{"declaration"},
						},
						Full: true,
					},
				},
				InlayHintProvider: &protocol.InlayHintOptions{},
			},
			ServerInfo: &protocol316.InitializeResultServerInfo{
				Name:    name,
				Version: &version,
			},
		}
		return capabilities, nil
	}
}

func initializedHandler() protocol316.InitializedFunc {
	return func(context *glsp.Context, params *protocol316.InitializedParams) error {
		return nil
	}
}

func shutdownHandler() protocol316.ShutdownFunc {
	return func(context *glsp.Context) error {
		return nil
	}
}

func setTraceHandler() protocol316.SetTraceFunc {
	return func(context *glsp.Context, params *protocol316.SetTraceParams) error {
		return nil
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func ptrTo[T any](v T) *T {
	return &v
}
