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
            | timer_stmt
            | hint_stmt
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

### Timer

```
timer_stmt ::= 'timer' duration NEWLINE
duration ::= NUMBER ('s' | 'm' | 'h' | 'd') | IDENT
```

### Hint Statement

```
hint_stmt ::= 'hint' ('signal' | 'update' | 'query') IDENT NEWLINE
```

Documents that a signal, update, or query handler may execute at this point. Hints are regular statements that can appear anywhere in the workflow.

**Common Pattern:** Place `hint` statements at the beginning of `watch` case bodies to document which signals/updates can affect the watched variable:

```
await one:
    watch (approved):
        hint signal Approved
        hint update AdminOverride
        close
```

**Note:** The IDENT must refer to a signal, query, or update declared in the current workflow. Each is declared once at the workflow level and can be referenced multiple times via hints throughout the workflow.

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

await_one_case ::= watch_case | timer_case | await_all_case

watch_case ::= 'watch' '(' IDENT ')' ':' NEWLINE
               INDENT statement* DEDENT

timer_case ::= 'timer' '(' duration ')' ':' NEWLINE
               INDENT statement* DEDENT

await_all_case ::= 'await' 'all' ':' NEWLINE
                   INDENT statement* DEDENT

duration ::= NUMBER ('s' | 'm' | 'h' | 'd') | IDENT
```

Waits for the FIRST case to complete (races between watch conditions, timers, and nested await all operations).

**Watch cases** wait for a state variable to become truthy. When the watched variable becomes true (typically set by a signal or update handler), the watch case body executes. Use `hint` statements within watch case bodies to document which signals/updates can affect the watched variable.

**Timer cases** wait for a duration to elapse. When the timer fires, the timer case body executes.

**Await all cases** wait for all statements in their body to complete. When all statements complete, the await all case wins.

The case that completes first "wins" the race, its body executes, and then execution continues after the `await one` block.

**Note:** Signals and updates modify state variables which are then watched. They cannot directly terminate workflows - only the main workflow body can call `close`.

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
- `await` - Wait for operations (`await all`, `await one`)
- `all` - Wait for all operations (used with `await`)
- `one` - Wait for first operation (used with `await`)

**Workflow primitives:**
- `workflow` - Workflow definition or child call
- `activity` - Activity definition or call
- `timer` - Durable sleep
- `signal` - Signal declaration (referenced via `hint`)
- `query` - Query declaration (referenced via `hint`)
- `update` - Update declaration (referenced via `hint`)
- `hint` - Mark where signal/update may arrive

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
        hint signal Done
    else:
        timer 1h

    return Result{status: status}
```

## Context Restrictions

### Temporal Keywords

Certain keywords are only valid in workflow context and produce errors in activity context:

- `spawn`, `detach`, `nexus` - Workflow calls
- `workflow` - Child workflow calls
- `timer` - Durable sleep
- `signal`, `query`, `update` - Handler declarations
- `await` - Signal/update waiting
- `hint` - Signal/update annotation
- `parallel` - Concurrent execution
- `select` - Racing branches
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
   - Resolve await targets to signal/update declarations
   - Resolve hint targets to signal/update declarations
   - Walk signal/query/update handler bodies and resolve references
3. **Report errors:** Undefined references, duplicate definitions, etc.

### Error Handling

The parser and resolver collect multiple errors before failing, allowing users to fix multiple issues in one pass.

Common error types:
- Undefined activity/workflow/signal/update
- Duplicate definitions
- Signal/update cases in select blocks (not allowed)
- Temporal keywords in activity context
- Hint referencing query (only signal/update allowed)

## Examples

See the `examples/` directory for complete working examples of all language features.

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

statement ::= activity_call | workflow_call | timer_stmt
            | hint_stmt | await_all_block | await_one_block | switch_block
            | if_stmt | for_stmt | close_stmt | return_stmt | continue_as_new_stmt
            | break_stmt | continue_stmt | assignment

await_one_case ::= watch_case | timer_case | await_all_case

watch_case ::= 'watch' '(' IDENT ')' ':' NEWLINE INDENT statement* DEDENT

timer_case ::= 'timer' '(' duration ')' ':' NEWLINE INDENT statement* DEDENT

close_stmt ::= 'close' ['completed' | 'failed'] [expr] NEWLINE
```
