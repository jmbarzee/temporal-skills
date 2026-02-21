package main

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/internal/server"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	glspServer "github.com/tliron/glsp/server"
)

// lspCommand starts the LSP server over stdio.
func lspCommand() {
	commonlog.Configure(1, nil)

	handler, _ := server.NewHandler(name, version)

	s := glspServer.NewServer(handler, name, false)

	s.RunStdio()
}
