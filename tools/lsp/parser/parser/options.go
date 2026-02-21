package parser

import (
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/ast"
	"github.com/jmbarzee/temporal-skills/tools/lsp/parser/token"
)

// OptionsContext identifies the parent context for schema validation.
type OptionsContext int

const (
	OptionsContextActivity OptionsContext = iota
	OptionsContextWorkflow
	OptionsContextWorker
	OptionsContextNexusCall
	OptionsContextEndpoint
)

// optionSchema describes the expected value type for an option key.
// "nested" means the key introduces a nested block (e.g. retry_policy:).
type optionSchema struct {
	valueType string // "string", "duration", "number", "bool", "enum", "nested"
	nested    map[string]*optionSchema
	allowed   []string // allowed values for enum type
}

var retryPolicySchema = map[string]*optionSchema{
	"initial_interval":        {valueType: "duration"},
	"backoff_coefficient":     {valueType: "number"},
	"maximum_interval":        {valueType: "duration"},
	"maximum_attempts":        {valueType: "number"},
	"non_retryable_error_types": {valueType: "string"},
}

var prioritySchema = map[string]*optionSchema{
	"value": {valueType: "number"},
}

var activityOptionSchema = map[string]*optionSchema{
	"task_queue":                    {valueType: "string"},
	"schedule_to_close_timeout":     {valueType: "duration"},
	"schedule_to_start_timeout":     {valueType: "duration"},
	"start_to_close_timeout":        {valueType: "duration"},
	"heartbeat_timeout":             {valueType: "duration"},
	"request_eager_execution":       {valueType: "bool"},
	"retry_policy":                  {valueType: "nested", nested: retryPolicySchema},
	"priority":                      {valueType: "nested", nested: prioritySchema},
}

var workerOptionSchema = map[string]*optionSchema{
	"task_queue":                                {valueType: "string"},
	"worker_activity_rate_limit":                {valueType: "number"},
	"task_queue_activity_rate_limit":            {valueType: "number"},
	"worker_local_activity_rate_limit":          {valueType: "number"},
	"max_concurrent_activity_executions":        {valueType: "number"},
	"max_concurrent_workflow_task_executions":   {valueType: "number"},
	"max_concurrent_local_activity_executions":  {valueType: "number"},
	"max_concurrent_workflow_task_pollers":      {valueType: "number"},
	"max_concurrent_activity_task_pollers":      {valueType: "number"},
	"max_cached_workflows":                      {valueType: "number"},
	"sticky_schedule_to_start_timeout":          {valueType: "duration"},
	"heartbeat_throttle_interval":               {valueType: "duration"},
	"worker_identity":                           {valueType: "string"},
	"worker_shutdown_timeout":                   {valueType: "duration"},
	"local_activity_only_mode":                  {valueType: "bool"},
}

var workflowOptionSchema = map[string]*optionSchema{
	"task_queue":                    {valueType: "string"},
	"workflow_execution_timeout":    {valueType: "duration"},
	"workflow_run_timeout":          {valueType: "duration"},
	"workflow_task_timeout":         {valueType: "duration"},
	"parent_close_policy":          {valueType: "enum", allowed: []string{"TERMINATE", "ABANDON", "REQUEST_CANCEL"}},
	"workflow_id_reuse_policy":     {valueType: "enum", allowed: []string{"ALLOW_DUPLICATE", "ALLOW_DUPLICATE_FAILED_ONLY", "REJECT_DUPLICATE", "TERMINATE_IF_RUNNING"}},
	"cron_schedule":                {valueType: "string"},
	"retry_policy":                 {valueType: "nested", nested: retryPolicySchema},
	"priority":                     {valueType: "nested", nested: prioritySchema},
}

var nexusCallOptionSchema = map[string]*optionSchema{
	"schedule_to_close_timeout": {valueType: "duration"},
	"retry_policy":              {valueType: "nested", nested: retryPolicySchema},
	"priority":                  {valueType: "nested", nested: prioritySchema},
}

var endpointOptionSchema = map[string]*optionSchema{
	"task_queue": {valueType: "string"},
}

func schemaForContext(ctx OptionsContext) map[string]*optionSchema {
	switch ctx {
	case OptionsContextActivity:
		return activityOptionSchema
	case OptionsContextWorkflow:
		return workflowOptionSchema
	case OptionsContextWorker:
		return workerOptionSchema
	case OptionsContextNexusCall:
		return nexusCallOptionSchema
	case OptionsContextEndpoint:
		return endpointOptionSchema
	default:
		return nil
	}
}

// parseOptionsBlock parses the contents of an options block: COLON NEWLINE INDENT entries DEDENT.
// The OPTIONS keyword has already been consumed. Expects current token = COLON.
func (p *Parser) parseOptionsBlock(ctx OptionsContext) (*ast.OptionsBlock, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}

	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if p.current.Type == token.NEWLINE {
		p.advance()
	}
	if _, err := p.expect(token.INDENT); err != nil {
		return nil, err
	}

	schema := schemaForContext(ctx)
	entries, err := p.parseOptionEntries(schema)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.DEDENT); err != nil {
		return nil, err
	}

	return &ast.OptionsBlock{
		Pos:     pos,
		Entries: entries,
	}, nil
}

// parseOptionEntries parses key-value pairs until DEDENT is encountered.
func (p *Parser) parseOptionEntries(schema map[string]*optionSchema) ([]*ast.OptionEntry, error) {
	var entries []*ast.OptionEntry

	for p.current.Type != token.DEDENT && p.current.Type != token.EOF {
		if p.current.Type == token.NEWLINE {
			p.advance()
			continue
		}

		entry, err := p.parseOptionEntry(schema)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// parseOptionEntry parses a single option entry: IDENT COLON (value | NEWLINE INDENT nested DEDENT).
func (p *Parser) parseOptionEntry(schema map[string]*optionSchema) (*ast.OptionEntry, error) {
	pos := ast.Pos{Line: p.current.Line, Column: p.current.Column}

	var key string
	switch p.current.Type {
	case token.IDENT:
		key = p.current.Literal
		p.advance()
	case token.TASK_QUEUE:
		key = "task_queue"
		p.advance()
	default:
		return nil, p.errorf("expected option key, got %s", p.current.Type)
	}

	// Look up schema for this key.
	var sch *optionSchema
	if schema != nil {
		var ok bool
		sch, ok = schema[key]
		if !ok {
			return nil, &ParseError{
				Msg:    "unknown option key: " + key,
				Line:   pos.Line,
				Column: pos.Column,
			}
		}
	}

	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}

	entry := &ast.OptionEntry{
		Pos: pos,
		Key: key,
	}

	// Branch: NEWLINE followed by INDENT means nested block, otherwise flat value.
	if p.current.Type == token.NEWLINE && p.peek.Type == token.INDENT {
		// Nested block.
		if sch != nil && sch.valueType != "nested" {
			return nil, &ParseError{
				Msg:    "option " + key + " expects a value, not a nested block",
				Line:   p.current.Line,
				Column: p.current.Column,
			}
		}

		p.advance() // consume NEWLINE
		p.advance() // consume INDENT

		var nestedSchema map[string]*optionSchema
		if sch != nil {
			nestedSchema = sch.nested
		}

		nested, err := p.parseOptionEntries(nestedSchema)
		if err != nil {
			return nil, err
		}
		entry.Nested = nested

		if _, err := p.expect(token.DEDENT); err != nil {
			return nil, err
		}
	} else {
		// Flat key: value.
		if sch != nil && sch.valueType == "nested" {
			return nil, &ParseError{
				Msg:    "option " + key + " expects a nested block, not a value",
				Line:   p.current.Line,
				Column: p.current.Column,
			}
		}

		value, valueType, err := p.parseOptionValue(sch)
		if err != nil {
			return nil, err
		}
		entry.Value = value
		entry.ValueType = valueType
	}

	// Consume trailing newline if present.
	if p.current.Type == token.NEWLINE {
		p.advance()
	}

	return entry, nil
}

// parseOptionValue parses a value after COLON. Returns value string and type.
func (p *Parser) parseOptionValue(sch *optionSchema) (string, string, error) {
	switch p.current.Type {
	case token.STRING:
		val := p.current.Literal
		p.advance()
		if sch != nil && sch.valueType != "string" {
			return "", "", &ParseError{
				Msg:    "expected " + sch.valueType + ", got string",
				Line:   p.current.Line,
				Column: p.current.Column,
			}
		}
		return val, "string", nil

	case token.DURATION:
		val := p.current.Literal
		p.advance()
		if sch != nil && sch.valueType != "duration" {
			return "", "", &ParseError{
				Msg:    "expected " + sch.valueType + ", got duration",
				Line:   p.current.Line,
				Column: p.current.Column,
			}
		}
		return val, "duration", nil

	case token.NUMBER:
		val := p.current.Literal
		p.advance()
		if sch != nil && sch.valueType != "number" {
			return "", "", &ParseError{
				Msg:    "expected " + sch.valueType + ", got number",
				Line:   p.current.Line,
				Column: p.current.Column,
			}
		}
		return val, "number", nil

	case token.IDENT:
		val := p.current.Literal
		p.advance()
		// Could be bool (true/false) or enum value.
		if val == "true" || val == "false" {
			if sch != nil && sch.valueType != "bool" {
				return "", "", &ParseError{
					Msg:    "expected " + sch.valueType + ", got bool",
					Line:   p.current.Line,
					Column: p.current.Column,
				}
			}
			return val, "bool", nil
		}
		// Enum value.
		if sch != nil && sch.valueType == "enum" {
			valid := false
			for _, a := range sch.allowed {
				if a == val {
					valid = true
					break
				}
			}
			if !valid {
				return "", "", &ParseError{
					Msg:    "invalid value " + val + " for enum option (allowed: " + joinStrings(sch.allowed) + ")",
					Line:   p.current.Line,
					Column: p.current.Column,
				}
			}
			return val, "enum", nil
		}
		// If no schema, treat as enum.
		return val, "enum", nil

	default:
		return "", "", p.errorf("expected value after colon, got %s", p.current.Type)
	}
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
