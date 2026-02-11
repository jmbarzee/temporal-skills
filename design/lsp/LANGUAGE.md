# TWF Language Specification

Formal specification for the Temporal Workflow Format (TWF) language.

## File Structure

A TWF file consists of zero or more top-level definitions:

```
file ::= definition*
definition ::= workflow_def | activity_def
```

## Workflow Definitions

```
workflow_def ::= 'workflow' IDENT params ['->' return_type] ':' NEWLINE
                 INDENT
                 [signal_decl*]
                 [query_decl*]
                 [update_decl*]
                 statement*
                 DEDENT

params ::= '(' [param_list] ')'
param_list ::= param (',' param)*
param ::= IDENT ':' type
return_type ::= '(' type_list ')'  # Always parenthesized
type_list ::= type (',' type)*
type ::= IDENT | type '[' type ']' | type '{' ... '}'
```

**Important:** Signal, query, and update declarations are optional but if present, must appear before workflow body statements. Each signal/query/update can only be declared once per workflow.

### Signal Declarations

Signal handlers are defined at the beginning of workflows with handler body blocks:

```
signal_decl ::= 'signal' IDENT params ':' NEWLINE
                INDENT
                statement*
                DEDENT
```

Signal handler bodies execute when the signal arrives. Handlers have access to the full workflow statement set (activities, child workflows, timers, etc.).

### Query Declarations

Query handlers are defined at the beginning of workflows with handler body blocks:

```
query_decl ::= 'query' IDENT params '->' return_type ':' NEWLINE
               INDENT
               statement*
               DEDENT
```

**Return type is required for queries** (always parenthesized, e.g., `-> (Status)`).

Query handler bodies are restricted to the activity statement set (no temporal primitives like timers, signals, or child workflows). Queries must not modify workflow state.

### Update Declarations

Update handlers are defined at the beginning of workflows with handler body blocks:

```
update_decl ::= 'update' IDENT params '->' return_type ':' NEWLINE
                INDENT
                statement*
                DEDENT
```

**Return type is required for updates** (always parenthesized, e.g., `-> (Result)`).

Update handler bodies execute when the update is received. Handlers have access to the full workflow statement set and can return values to the caller.

## Activity Definitions

```
activity_def ::= 'activity' IDENT params ['->' return_type] ':' NEWLINE
                 INDENT
                 [options_stmt]
                 statement*
                 DEDENT
```

Return type is optional; if present, must be parenthesized (e.g., `-> (Result)`).

Activities have access to a restricted statement set (no temporal primitives like timers or child workflows). Activities may use the `heartbeat()` primitive to report progress during long-running operations.

## Statements

### Workflow Statements

Available in workflow context (workflow definitions and signal/update handlers):

```
statement ::= activity_call
            | workflow_call
            | await_stmt
            | await_all_block
            | await_one_block
            | switch_block
            | if_stmt
            | for_stmt
            | close_stmt
            | return_stmt
            | continue_as_new_stmt
            | break_stmt
            | continue_stmt
            | assignment
```

### Activity Statements

Available in activity context (activity definitions and query handlers):

```
statement ::= heartbeat_stmt
            | switch_block
            | if_stmt
            | for_stmt
            | return_stmt
            | break_stmt
            | continue_stmt
            | assignment
```

## Statement Syntax

### Activity Call

```
activity_call ::= 'activity' IDENT args ['->' result] [NEWLINE options_block]

args ::= '(' [arg_list] ')'
arg_list ::= expr (',' expr)*
result ::= IDENT | '(' IDENT (',' IDENT)* ')'

options_block ::= INDENT 'options' '(' option_list ')' NEWLINE DEDENT
option_list ::= option (',' option)*
option ::= IDENT ':' expr
```

**Note:** When using options blocks, the `options(...)` line must be indented on the line following the activity call. Options blocks require arrow syntax (`activity Foo() -> result`) not assignment syntax (`result = activity Foo()`).

### Workflow Call

```
workflow_call ::= ['spawn' | 'detach'] ['nexus' STRING] 'workflow' IDENT args ['->' result] [NEWLINE options_block]
```

Modifiers:
- `spawn`: Asynchronous child workflow (returns handle)
- `detach`: Fire-and-forget child workflow (no result)
- `nexus STRING`: Cross-namespace workflow call

### Single Await Statement

```
await_stmt ::= 'await' await_target NEWLINE

await_target ::= timer_target
               | signal_target
               | update_target
               | activity_target
               | workflow_target

timer_target ::= 'timer' '(' duration ')'

signal_target ::= 'signal' IDENT ['->' params]

update_target ::= 'update' IDENT ['->' params]

activity_target ::= 'activity' IDENT args ['->' result]

workflow_target ::= ['spawn' | 'detach'] ['nexus' STRING] 'workflow' IDENT args ['->' result]

duration ::= NUMBER ('s' | 'm' | 'h' | 'd') | IDENT
params ::= '(' IDENT (',' IDENT)* ')'
result ::= IDENT | '(' IDENT (',' IDENT)* ')'
```

Single await blocks until the specified operation completes. For signals and updates, the handler body executes first, then the await continues. For activities and workflows, the result is bound to the specified variable(s).

**Examples:**
```
await timer(5m)
await signal Approved
await signal Approved -> (approver, timestamp)
await activity Process(data) -> result
await workflow Child(input) -> output
```

### Await All Block

```
await_all_block ::= 'await' 'all' ':' NEWLINE
                    INDENT
                    statement*
                    DEDENT
```

Executes all contained statements concurrently and waits for ALL to complete before continuing.

### Await One Block

```
await_one_block ::= 'await' 'one' ':' NEWLINE
                    INDENT
                    await_one_case+
                    DEDENT

await_one_case ::= signal_case
                 | update_case
                 | timer_case
                 | activity_case
                 | workflow_case
                 | await_all_case

signal_case ::= 'signal' IDENT ['->' params] ':' NEWLINE
                [INDENT statement+ DEDENT]

update_case ::= 'update' IDENT ['->' params] ':' NEWLINE
                [INDENT statement+ DEDENT]

timer_case ::= 'timer' '(' duration ')' ':' NEWLINE
               [INDENT statement+ DEDENT]

activity_case ::= 'activity' IDENT args ['->' result] ':' NEWLINE
                  [INDENT statement+ DEDENT]

workflow_case ::= ['spawn' | 'detach'] ['nexus' STRING] 'workflow' IDENT args ['->' result] ':' NEWLINE
                  [INDENT statement+ DEDENT]

await_all_case ::= 'await' 'all' ':' NEWLINE
                   INDENT statement+ DEDENT

duration ::= NUMBER ('s' | 'm' | 'h' | 'd') | IDENT
params ::= '(' IDENT (',' IDENT)* ')'
result ::= IDENT | '(' IDENT (',' IDENT)* ')'
```

Waits for the FIRST case to complete (races between signals, updates, timers, activities, workflows, and nested await all operations).

**Signal cases** wait for a specific signal to arrive. When the signal arrives, the handler body executes first (if defined), then the case body executes (if present). Signal parameters can be bound using `->`.

**Update cases** wait for a specific update to arrive. When the update arrives, the handler body executes and returns a value to the caller, then the case body executes (if present). Update parameters can be bound using `->`.

**Timer cases** wait for a duration to elapse. When the timer fires, the case body executes (if present).

**Activity cases** wait for an activity to complete. When the activity completes, the case body executes (if present). Activity results can be bound using `->`.

**Workflow cases** wait for a child workflow to complete. When the workflow completes, the case body executes (if present). Workflow results can be bound using `->`.

**Await all cases** wait for all statements in their body to complete. When all statements complete, the await all case wins.

**Case bodies are optional.** If a case has no body, the colon is still required. This is useful for consuming signals/results without additional processing:
```
await one:
    signal Ready:
    timer(5m):
        close failed "timeout"
```

The case that completes first "wins" the race, its body executes (if present), and then execution continues after the `await one` block.

**Cancellation:** When one case completes, all other pending operations are automatically cancelled. Activities receive cancellation signals, child workflows are cancelled, and timers are stopped.

### Switch Block

```
switch_block ::= 'switch' '(' expr ')' ':' NEWLINE
                 INDENT
                 switch_case+
                 [else_case]
                 DEDENT

switch_case ::= 'case' expr ':' NEWLINE
                INDENT statement* DEDENT

else_case ::= 'else' ':' NEWLINE
              INDENT statement* DEDENT
```

### If Statement

```
if_stmt ::= 'if' '(' expr ')' ':' NEWLINE
            INDENT statement* DEDENT
            ['else' ':' NEWLINE INDENT statement* DEDENT]
```

### For Statement

```
for_stmt ::= 'for' [for_header] ':' NEWLINE
             INDENT statement* DEDENT

for_header ::= '(' expr ')' | '(' IDENT 'in' expr ')'
```

- No header: infinite loop
- `(expr)`: conditional loop (while expr)
- `(item in items)`: iteration loop

### Close Statement

```
close_stmt ::= 'close' [close_reason] [expr] NEWLINE
close_reason ::= 'completed' | 'failed'
```

Terminates workflow execution with a completion status. Only valid in workflow context (not in activities or queries).

- `close` or `close completed` - Normal successful completion
- `close failed` - Terminates workflow in failed state, optionally with error message/data

**Important:** Signals and updates cannot call `close` - they can only mutate state. Only the main workflow body can terminate execution using `close`.

**Note:** `return` is still valid in queries (which must return values without terminating the workflow) and can be used in workflows for backward compatibility, but `close` is preferred for workflow termination as it makes the intent explicit.

### Return Statement

```
return_stmt ::= 'return' [expr] NEWLINE
```

Used in queries and activities to return values. In workflows, prefer `close` for termination.

### Continue As New

```
continue_as_new_stmt ::= 'continue_as_new' '(' arg_list ')' NEWLINE
```

Resets workflow history and continues with new arguments. Used for long-running workflows.

### Break and Continue

```
break_stmt ::= 'break' NEWLINE
continue_stmt ::= 'continue' NEWLINE
```

### Assignment

```
assignment ::= IDENT '=' expr NEWLINE
```

### Heartbeat (Activity-only)

```
heartbeat_stmt ::= 'heartbeat' '(' [arg_list] ')' NEWLINE
```

The `heartbeat()` primitive is only available in activity bodies. It reports progress to the Temporal service, allowing activities to be resumed if they fail mid-execution. Optional arguments can include progress details.

## Expressions

```
expr ::= IDENT
       | NUMBER
       | STRING
       | 'true' | 'false'
       | 'null'
       | binary_expr
       | unary_expr
       | call_expr
       | index_expr
       | field_expr
       | constructor_expr

binary_expr ::= expr binary_op expr
binary_op ::= '+' | '-' | '*' | '/' | '%'
            | '==' | '!=' | '<' | '<=' | '>' | '>='
            | 'and' | 'or'

unary_expr ::= unary_op expr
unary_op ::= '-' | 'not'

call_expr ::= IDENT '(' [arg_list] ')'
index_expr ::= expr '[' expr ']'
field_expr ::= expr '.' IDENT
constructor_expr ::= IDENT '{' [field_list] '}'
field_list ::= field (',' field)*
field ::= IDENT ':' expr
```

## Tokens and Keywords

### Keywords

**Async workflow operations:**
- `spawn` - Asynchronous child workflow
- `detach` - Fire-and-forget child workflow
- `nexus` - Cross-namespace call
- `await` - Wait for operations (`await timer`, `await signal`, `await all`, `await one`)
- `all` - Wait for all operations (used with `await`)
- `one` - Wait for first operation (used with `await`)

**Workflow primitives:**
- `workflow` - Workflow definition or child call
- `activity` - Activity definition or call
- `timer` - Durable sleep (used with `await`)
- `signal` - Signal declaration and await target
- `query` - Query declaration
- `update` - Update declaration and await target

**Activity primitives:**
- `heartbeat` - Report activity progress (activity-only)

**Control flow:**
- `switch` - Multi-way conditional
- `case` - Switch case
- `if` - Conditional
- `else` - Alternative branch
- `for` - Loop
- `in` - Iteration operator

**Flow control:**
- `return` - Return from definition
- `continue_as_new` - Reset history and continue
- `break` - Exit loop
- `continue` - Next loop iteration

**Operators:**
- `and`, `or`, `not` - Logical operators

**Configuration:**
- `options` - Options block for calls/definitions

### Identifiers

```
IDENT ::= [a-zA-Z_][a-zA-Z0-9_]*
```

Identifiers start with a letter or underscore, followed by any combination of letters, digits, or underscores.

### Literals

```
NUMBER ::= [0-9]+(['.'[0-9]+])
STRING ::= '"' [^"]* '"'
```

### Comments

```
comment ::= '#' .* NEWLINE
```

Comments start with `#` and continue to the end of the line. Comments can appear anywhere in the source and are captured in the AST but do not affect execution semantics.

## Indentation Rules

TWF uses **indentation-based scoping** (like Python):

1. **Consistent indentation:** Use either tabs or spaces consistently throughout a file
2. **Block start:** A colon (`:`) followed by NEWLINE and INDENT starts a new block
3. **Block end:** DEDENT ends the current block
4. **Blank lines:** Blank lines (with or without whitespace) are skipped
5. **No mixing:** Do not mix tabs and spaces in the same file

### Example

```
workflow Example(x: int) -> (Result):
    signal Done():
        status = "done"

    if (x > 0):
        activity DoWork(x)
    else:
        await timer(1h)

    return Result{status: status}
```

## Context Restrictions

### Temporal Keywords

Certain keywords are only valid in workflow context and produce errors in activity context:

- `spawn`, `detach`, `nexus` - Workflow calls
- `workflow` - Child workflow calls
- `timer` - Durable sleep (with `await`)
- `signal`, `query`, `update` - Handler declarations and await targets
- `await` - Async operation waiting
- `continue_as_new` - History reset

These keywords are **blocked in:**
- Activity definitions
- Query handler bodies

### Handler Body Contexts

- **Signal handlers:** Full workflow statement set, but cannot call `close` (can only mutate state)
- **Update handlers:** Full workflow statement set, but cannot call `close` (can only mutate state)
- **Query handlers:** Activity statement set (no temporal primitives), use `return` for values

## Semantic Rules

### Resolution

After parsing, the resolver performs symbol resolution:

1. **Build symbol table:** Collect all workflow and activity definitions
2. **Per-workflow resolution:**
   - Build signal/query/update maps for the workflow
   - Resolve activity calls to activity definitions
   - Resolve workflow calls to workflow definitions
   - Resolve await targets to signal/update/activity/workflow declarations
   - Walk signal/query/update handler bodies and resolve references
3. **Report errors:** Undefined references, duplicate definitions, etc.

### Error Handling

The parser and resolver collect multiple errors before failing, allowing users to fix multiple issues in one pass.

Common error types:
- Undefined activity/workflow/signal/update
- Duplicate definitions
- Temporal keywords in activity context
- Invalid await targets (e.g., awaiting a query)

## Examples

See the `topics/` directory for complete working examples of all language features.

## Grammar Summary

```
file ::= definition*
definition ::= workflow_def | activity_def

workflow_def ::= 'workflow' IDENT params ['->' return_type] ':'
                 NEWLINE INDENT
                 [signal_decl*] [query_decl*] [update_decl*]
                 statement*
                 DEDENT

activity_def ::= 'activity' IDENT params ['->' return_type] ':'
                 NEWLINE INDENT statement* DEDENT

signal_decl ::= 'signal' IDENT params ':' NEWLINE INDENT statement* DEDENT
query_decl ::= 'query' IDENT params '->' return_type ':' NEWLINE INDENT statement* DEDENT
update_decl ::= 'update' IDENT params '->' return_type ':' NEWLINE INDENT statement* DEDENT

statement ::= activity_call | workflow_call | await_stmt
            | await_all_block | await_one_block | switch_block
            | if_stmt | for_stmt | close_stmt | return_stmt | continue_as_new_stmt
            | break_stmt | continue_stmt | assignment

await_stmt ::= 'await' (timer_target | signal_target | update_target | activity_target | workflow_target) NEWLINE

await_one_case ::= signal_case | update_case | timer_case | activity_case | workflow_case | await_all_case

signal_case ::= 'signal' IDENT ['->' params] ':' NEWLINE [INDENT statement+ DEDENT]

update_case ::= 'update' IDENT ['->' params] ':' NEWLINE [INDENT statement+ DEDENT]

timer_case ::= 'timer' '(' duration ')' ':' NEWLINE [INDENT statement+ DEDENT]

activity_case ::= 'activity' IDENT args ['->' result] ':' NEWLINE [INDENT statement+ DEDENT]

workflow_case ::= ['spawn' | 'detach'] ['nexus' STRING] 'workflow' IDENT args ['->' result] ':' NEWLINE [INDENT statement+ DEDENT]

close_stmt ::= 'close' ['completed' | 'failed'] [expr] NEWLINE
```
