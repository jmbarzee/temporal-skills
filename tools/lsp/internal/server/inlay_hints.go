package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_17"
)

// TODO: implement inlay hints (timer durations, resolved types, etc.)
func inlayHintHandler(store *DocumentStore) protocol.TextDocumentInlayHintFunc {
	return func(context *glsp.Context, params *protocol.InlayHintParams) ([]protocol.InlayHint, error) {
		return nil, nil
	}
}
