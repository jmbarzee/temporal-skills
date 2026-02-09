package main

import (
	"github.com/jmbarzee/temporal-skills/lsp/internal/server"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	glspServer "github.com/tliron/glsp/server"
)

const (
	name    = "twf-lsp"
	version = "0.1.0"
)

func main() {
	commonlog.Configure(1, nil)

	handler := server.NewHandler(name, version)

	s := glspServer.NewServer(handler, name, false)

	s.RunStdio()
}
