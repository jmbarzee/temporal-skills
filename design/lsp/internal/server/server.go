package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// NewHandler creates a protocol.Handler with all LSP methods registered.
func NewHandler(name, version string) *protocol.Handler {
	store := NewDocumentStore()

	handler := &protocol.Handler{
		Initialize:  initializeHandler(name, version),
		Initialized: initializedHandler(),
		Shutdown:    shutdownHandler(),
		SetTrace:    setTraceHandler(),

		TextDocumentDidOpen:  didOpenHandler(store),
		TextDocumentDidChange: didChangeHandler(store),
		TextDocumentDidClose: didCloseHandler(store),

		TextDocumentHover:          hoverHandler(store),
		TextDocumentDefinition:     definitionHandler(store),
		TextDocumentDocumentSymbol: documentSymbolHandler(store),
		TextDocumentCompletion:     completionHandler(store),
		TextDocumentReferences:     referencesHandler(store),
		TextDocumentRename:         renameHandler(store),
		TextDocumentPrepareRename:  prepareRenameHandler(store),
	}

	return handler
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
				ReferencesProvider: &protocol.ReferenceOptions{},
				RenameProvider:     &protocol.RenameOptions{PrepareProvider: boolPtr(true)},
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
