package server

import (
	"sync"

	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/parser"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/resolver"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/validator"
)

// Document holds the content and analysis results for a single open file.
type Document struct {
	URI          string
	Content      string
	File         *ast.File
	ParseErrs    []*parser.ParseError
	ResolveErrs  []*resolver.ResolveError
	ValidateErrs []*validator.Error
}

// analyze parses, resolves, and validates the document content.
func (d *Document) analyze() {
	d.File = nil
	d.ParseErrs = nil
	d.ResolveErrs = nil
	d.ValidateErrs = nil

	f, errs := parser.ParseFileAll(d.Content)
	d.File = f
	d.ParseErrs = errs

	if len(f.Definitions) > 0 {
		d.ResolveErrs = resolver.Resolve(f)
		d.ValidateErrs = validator.Validate(f)
	}
}

// DocumentStore is a thread-safe store of open documents.
type DocumentStore struct {
	mu   sync.RWMutex
	docs map[string]*Document
}

// NewDocumentStore creates an empty document store.
func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		docs: make(map[string]*Document),
	}
}

// Open adds or replaces a document in the store and analyzes it.
func (s *DocumentStore) Open(uri, content string) *Document {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc := &Document{URI: uri, Content: content}
	doc.analyze()
	s.docs[uri] = doc
	return doc
}

// Update updates the content of an existing document and re-analyzes it.
func (s *DocumentStore) Update(uri, content string) *Document {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, ok := s.docs[uri]
	if !ok {
		doc = &Document{URI: uri}
		s.docs[uri] = doc
	}
	doc.Content = content
	doc.analyze()
	return doc
}

// Get returns a document by URI.
func (s *DocumentStore) Get(uri string) (*Document, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doc, ok := s.docs[uri]
	return doc, ok
}

// Close removes a document from the store.
func (s *DocumentStore) Close(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.docs, uri)
}
