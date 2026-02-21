# TWF Language Specification

Formal specification for the Temporal Workflow Format (TWF) language.

## File Structure

A TWF file consists of zero or more top-level definitions:

```
file ::= definition*
definition ::= workflow_def | activity_def | worker_def | namespace_def | nexus_service_def
```

## Workflow Definitions

```
workflow_def ::= 'workflow' IDENT params ['->' return_type] ':' NEWLINE
                 INDENT
                 [state_block]
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

**Important:** The state block (if present) must appear first, followed by signal/query/update declarations, then body statements. Each signal/query/update can only be declared once per workflow.

### State Block

The state block declares workflow state including named conditions and variable initializations. It must appear before signal/query/update declarations:

```
state_block ::= 'state' ':' NEWLINE
                INDENT
                state_stmt*
                DEDENT

state_stmt ::= condition_decl | raw_stmt

condition_decl ::= 'condition' IDENT NEWLINE
```

**Restrictions:** No temporal primitives inside `state:` block. It is purely declarative.

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
                 statement*
                 DEDENT
```

Return type is optional; if present, must be parenthesized (e.g., `-> (Result)`).

Activities have access to a restricted statement set (no temporal primitives like timers or child workflows). Activities may use the `heartbeat()` primitive to report progress during long-running operations.

## Worker Definitions

Workers are reusable type sets that group workflows and activities:

```
worker_def ::= 'worker' IDENT ':' NEWLINE
               INDENT
               worker_entry*
               DEDENT

worker_entry ::= 'workflow' IDENT NEWLINE
               | 'activity' IDENT NEWLINE
               | 'nexus' 'service' IDENT NEWLINE
```

Worker names use lowerCamelCase convention. Workers contain workflow, activity, and nexus service references — deployment configuration (task_queue, etc.) is specified when the worker is instantiated in a namespace block.

**Example:**
```
worker orderTypes:
    workflow ProcessOrder
    workflow CancelOrder
    activity ChargePayment
    activity SendNotification
    nexus service OrderService
```

## Namespace Definitions

Namespaces instantiate workers with deployment options, defining the deployment topology:

```
namespace_def ::= 'namespace' IDENT ':' NEWLINE
                  INDENT
                  namespace_entry*
                  DEDENT

namespace_entry ::= 'worker' IDENT NEWLINE [options_line]
                  | 'nexus' 'endpoint' IDENT NEWLINE [options_line]
```

Each worker instantiation inside a namespace requires a `task_queue` option. Nexus endpoint instantiations also require a `task_queue` option for routing.

**Example:**
```
namespace orders:
    worker orderTypes
        options:
            task_queue: "orderProcessing"
            max_concurrent_activity_executions: 50
    nexus endpoint OrderEndpoint
        options:
            task_queue: "orderProcessing"
```

The same worker type set can be instantiated in multiple namespaces with different options:

```
namespace staging:
    worker orderTypes
        options:
            task_queue: "staging-orders"
```

### Worker Options

Worker instantiation options (all snake_case):

| Key | Type |
|-----|------|
| `task_queue` | string (required) |
| `worker_activity_rate_limit` | number |
| `task_queue_activity_rate_limit` | number |
| `worker_local_activity_rate_limit` | number |
| `max_concurrent_activity_executions` | number |
| `max_concurrent_workflow_task_executions` | number |
| `max_concurrent_local_activity_executions` | number |
| `max_concurrent_workflow_task_pollers` | number |
| `max_concurrent_activity_task_pollers` | number |
| `max_cached_workflows` | number |
| `sticky_schedule_to_start_timeout` | duration |
| `heartbeat_throttle_interval` | duration |
| `worker_identity` | string |
| `worker_shutdown_timeout` | duration |
| `local_activity_only_mode` | bool |

### Endpoint Options

Nexus endpoint instantiation options:

| Key | Type |
|-----|------|
| `task_queue` | string (required) |

### Resolution

The resolver validates workers, namespaces, and nexus definitions:
- Worker references to undefined workflows, activities, or nexus services produce errors
- Duplicate worker, namespace, or nexus service names produce errors
- Duplicate nexus endpoint names across namespaces produce errors
- Namespace references to undefined workers produce errors
- Worker instantiations missing `task_queue` option produce errors
- Nexus endpoint instantiations missing `task_queue` option produce errors
- Workers on the same task queue (within a namespace) with different type sets produce errors
- Workers on the same task queue with identical type sets produce warnings (redundant)
- Nexus endpoint routing to a task queue where no worker registers the service produces errors
- Defined workflows/activities not on any instantiated worker produce warnings (when namespaces exist)
- Defined nexus services not referenced by any worker produce warnings (when namespaces exist)
- Defined workers not instantiated in any namespace produce warnings (when namespaces exist)

## Nexus Service Definitions

Nexus services define typed operation groups for cross-namespace communication:

```
nexus_service_def ::= 'nexus' 'service' IDENT ':' NEWLINE
                      INDENT nexus_operation* DEDENT

nexus_operation ::= async_operation | sync_operation

async_operation ::= 'async' IDENT 'workflow' IDENT NEWLINE

sync_operation  ::= 'sync' IDENT params '->' return_type ':' NEWLINE
                    INDENT statement* DEDENT
```

- `service` is a soft keyword (IDENT checked contextually after `nexus`)
- `sync` and `async` are hard keyword tokens
- **Async operations** delegate to a named workflow (one-liner, no body)
- **Sync operations** have a body using the workflow statement set (activities, queries, control flow, close)

**Example:**
```
nexus service OrderService:
    async PlaceOrder workflow ProcessOrder
    sync GetStatus(orderId: string) -> (Status):
        activity FetchStatus(orderId) -> status
        close complete(status)
```

### Resolution

The resolver validates nexus service definitions:
- Duplicate nexus service names produce errors
- Async operations referencing undefined workflows produce errors
- Sync operation bodies are resolved like workflow bodies

## Statements

### Workflow Statements

Available in workflow context (workflow definitions and signal/update handlers):

```
statement ::= activity_call
            | workflow_call
            | nexus_call
            | promise_stmt
            | set_stmt
            | unset_stmt
            | await_stmt
            | await_all_block
            | await_one_block
            | switch_block
            | if_stmt
            | for_stmt
            | close_stmt
            | return_stmt
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
activity_call ::= 'activity' IDENT args ['->' result] [NEWLINE options_line]

args ::= '(' [arg_list] ')'
arg_list ::= expr (',' expr)*
result ::= IDENT | '(' IDENT (',' IDENT)* ')'

options_line ::= INDENT 'options' ':' NEWLINE INDENT option_entry+ DEDENT NEWLINE DEDENT
```

**Note:** When using options blocks, the `options:` block must be indented on the line following the activity call.

### Options Block

```
options_block ::= 'options' ':' NEWLINE INDENT option_entry+ DEDENT
option_entry  ::= IDENT ':' value NEWLINE
                | IDENT ':' NEWLINE INDENT option_entry+ DEDENT

value ::= STRING | DURATION | NUMBER | IDENT

DURATION ::= NUMBER ('ms' | 's' | 'm' | 'h' | 'd')
NUMBER ::= [0-9]+ ['.' [0-9]+]
```

Options blocks use indentation-based nesting (same as the rest of TWF). Each key-value pair goes on its own line. Nested blocks (like `retry_policy`) use deeper indentation.

**Allowed keys per context:**

Activity call options: `task_queue`, `schedule_to_close_timeout`, `schedule_to_start_timeout`, `start_to_close_timeout`, `heartbeat_timeout`, `request_eager_execution`, `retry_policy`, `priority`

Workflow call options: `task_queue`, `workflow_execution_timeout`, `workflow_run_timeout`, `workflow_task_timeout`, `parent_close_policy`, `workflow_id_reuse_policy`, `cron_schedule`, `retry_policy`, `priority`

Retry policy keys: `initial_interval`, `backoff_coefficient`, `maximum_interval`, `maximum_attempts`, `non_retryable_error_types`

Priority keys: `priority_key` (number, 1–n, lower = higher priority), `fairness_key` (string, fairness balancing key), `fairness_weight` (number, weight in [0.001, 1000])

**Example:**
```
activity ChargePayment(order) -> payment
    options:
        task_queue: "payment-workers"
        start_to_close_timeout: 60s
        retry_policy:
            maximum_attempts: 3
            initial_interval: 1s
        priority:
            priority_key: 1
            fairness_key: "high"
```

### Workflow Call

```
workflow_call ::= ['detach'] 'workflow' IDENT args ['->' result] [NEWLINE options_line]
```

Modifiers:
- `detach`: Fire-and-forget child workflow (no result)

### Nexus Call

```
nexus_call ::= ['detach'] 'nexus' IDENT IDENT '.' IDENT args ['->' result] [NEWLINE options_line]
```

Calls a nexus service operation. The three IDENTs are: Endpoint, Service.Operation (dot-separated).

- `detach`: Fire-and-forget nexus call (no result)
- Endpoint: The nexus endpoint name (defined in a namespace block)
- Service.Operation: The nexus service and operation name (dot-separated)

**Nexus call options:** `schedule_to_close_timeout`, `retry_policy`, `priority` (all nested blocks use the same sub-key schemas described above)

**Examples:**
```
nexus OrderEndpoint OrderService.PlaceOrder(order) -> result
nexus OrderEndpoint OrderService.GetStatus(order.id) -> status
    options:
        schedule_to_close_timeout: 30s
detach nexus NotificationEndpoint NotificationService.SendEmail(email)
```

### Promise Statement

```
promise_stmt ::= 'promise' IDENT '<-' async_target NEWLINE

async_target ::= timer_target
               | signal_target
               | update_target
               | activity_target
               | workflow_target
               | nexus_target

nexus_target ::= 'nexus' IDENT IDENT '.' IDENT args
```

Declares a non-blocking async operation. The `<-` operator visually distinguishes async declaration from sync result binding (`->`). Use `await` to wait for the promise later.

**Examples:**
```
promise p <- activity ProcessItem(input)
promise report <- workflow BuildReport(data)
promise timeout <- timer(5m)
promise approved <- signal Approved
promise addr <- update ChangeAddress
promise result <- nexus OrderEndpoint OrderService.PlaceOrder(order)
```

### Set / Unset Statements

```
set_stmt ::= 'set' IDENT NEWLINE
unset_stmt ::= 'unset' IDENT NEWLINE
```

Set or unset a named condition declared in the workflow's `state:` block. Conditions can be awaited or used in `await one` cases.

**Examples:**
```
set clusterStarted
unset clusterStarted
```

### Single Await Statement

```
await_stmt ::= 'await' await_target NEWLINE

await_target ::= timer_target
               | signal_target
               | update_target
               | activity_target
               | workflow_target
               | nexus_target
               | ident_target

timer_target ::= 'timer' '(' duration ')'

signal_target ::= 'signal' IDENT ['->' params]

update_target ::= 'update' IDENT ['->' params]

activity_target ::= 'activity' IDENT args ['->' result]

workflow_target ::= ['detach'] 'workflow' IDENT args ['->' result]

nexus_target ::= ['detach'] 'nexus' IDENT IDENT '.' IDENT args ['->' result]

ident_target ::= IDENT ['->' result]

duration ::= NUMBER ('s' | 'm' | 'h' | 'd') | IDENT
params ::= '(' IDENT (',' IDENT)* ')'
result ::= IDENT | '(' IDENT (',' IDENT)* ')'
```

Single await blocks until the specified operation completes. For signals and updates, the handler body executes first, then the await continues. For activities and workflows, the result is bound to the specified variable(s). For ident targets, the name must refer to a previously declared promise or condition.

**Examples:**
```
await timer(5m)
await signal Approved
await signal Approved -> (approver, timestamp)
await activity Process(data) -> result
await workflow Child(input) -> output
await nexus OrderEndpoint OrderService.GetStatus(id) -> status
await myPromise -> result
await clusterStarted
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
                 | nexus_case
                 | await_all_case
                 | ident_case

signal_case ::= 'signal' IDENT ['->' params] ':' NEWLINE
                [INDENT statement+ DEDENT]

update_case ::= 'update' IDENT ['->' params] ':' NEWLINE
                [INDENT statement+ DEDENT]

timer_case ::= 'timer' '(' duration ')' ':' NEWLINE
               [INDENT statement+ DEDENT]

activity_case ::= 'activity' IDENT args ['->' result] ':' NEWLINE
                  [INDENT statement+ DEDENT]

workflow_case ::= ['detach'] 'workflow' IDENT args ['->' result] ':' NEWLINE
                  [INDENT statement+ DEDENT]

nexus_case ::= ['detach'] 'nexus' IDENT IDENT '.' IDENT args ['->' result] ':' NEWLINE
               [INDENT statement+ DEDENT]

await_all_case ::= 'await' 'all' ':' NEWLINE
                   INDENT statement+ DEDENT

ident_case ::= IDENT ['->' result] ':' NEWLINE
               [INDENT statement+ DEDENT]

duration ::= NUMBER ('s' | 'm' | 'h' | 'd') | IDENT
params ::= '(' IDENT (',' IDENT)* ')'
result ::= IDENT | '(' IDENT (',' IDENT)* ')'
```

Waits for the FIRST case to complete (races between signals, updates, timers, activities, workflows, promises, conditions, and nested await all operations).

**Signal cases** wait for a specific signal to arrive. When the signal arrives, the handler body executes first (if defined), then the case body executes (if present). Signal parameters can be bound using `->`.

**Update cases** wait for a specific update to arrive. When the update arrives, the handler body executes and returns a value to the caller, then the case body executes (if present). Update parameters can be bound using `->`.

**Timer cases** wait for a duration to elapse. When the timer fires, the case body executes (if present).

**Activity cases** wait for an activity to complete. When the activity completes, the case body executes (if present). Activity results can be bound using `->`.

**Workflow cases** wait for a child workflow to complete. When the workflow completes, the case body executes (if present). Workflow results can be bound using `->`.

**Await all cases** wait for all statements in their body to complete. When all statements complete, the await all case wins.

**Ident cases** wait for a named promise to resolve or a named condition to become true. Promise cases may bind a result using `->`. Condition cases cannot have `-> result` bindings. The name must refer to a previously declared promise or condition.

**Case bodies are optional.** If a case has no body, the colon is still required. This is useful for consuming signals/results without additional processing:
```
await one:
    signal Ready:
    timer(5m):
        close fail("timeout")
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
close_stmt ::= 'close' ('complete' | 'fail' | 'continue_as_new') ['(' args ')'] NEWLINE
```

Terminates workflow execution with an explicit exit state. Only valid in workflow context (not in activities or queries).

- `close complete` - Normal successful completion
- `close complete(Result{...})` - Completion with a return value
- `close fail` - Terminates workflow in failed state
- `close fail(Error{...})` - Failure with error data
- `close continue_as_new(args)` - Resets workflow history and continues with new arguments (for long-running workflows)

**Important:** Signals and updates cannot call `close` - they can only mutate state. Only the main workflow body can terminate execution using `close`.

**Note:** `return` is still valid in queries (which must return values without terminating the workflow) and can be used in workflows for backward compatibility, but `close` is preferred for workflow termination as it makes the intent explicit.

### Return Statement

```
return_stmt ::= 'return' [expr] NEWLINE
```

Used in queries and activities to return values. In workflows, prefer `close` for termination.

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
- `promise` - Declare a non-blocking async operation (binds with `<-`)
- `detach` - Fire-and-forget child workflow or nexus call
- `nexus` - Nexus service definition (top-level) or nexus call (in workflow body)
- `await` - Wait for operations (`await timer`, `await signal`, `await all`, `await one`, `await <promise>`, `await <condition>`)
- `all` - Wait for all operations (used with `await`)
- `one` - Wait for first operation (used with `await`)

**Workflow primitives:**
- `workflow` - Workflow definition or child call
- `activity` - Activity definition or call
- `timer` - Durable sleep (used with `await`)
- `signal` - Signal declaration and await target
- `query` - Query declaration
- `update` - Update declaration and await target

**State and conditions:**
- `state` - Workflow state declaration block
- `condition` - Named boolean awaitable (declared in `state:` block)
- `set` - Set a condition to true
- `unset` - Set a condition to false

**Activity primitives:**
- `heartbeat` - Report activity progress (activity-only)

**Control flow:**
- `switch` - Multi-way conditional
- `case` - Switch case
- `if` - Conditional
- `else` - Alternative branch
- `for` - Loop
- `in` - Iteration operator

**Workflow termination:**
- `close` - Terminate workflow execution
- `complete` - Successful completion (used with `close`)
- `fail` - Failed completion (used with `close`)
- `continue_as_new` - Reset history and continue (used with `close`)

**Flow control:**
- `return` - Return from definition
- `break` - Exit loop
- `continue` - Next loop iteration

**Operators:**
- `and`, `or`, `not` - Logical operators

**Nexus operations:**
- `sync` - Synchronous nexus operation (in nexus service body)
- `async` - Asynchronous nexus operation (in nexus service body)

**Worker topology:**
- `worker` - Worker type set definition (at top level) or worker instantiation (in namespace block)
- `namespace` - Namespace definition (deployment topology)
- `task_queue` - Task queue option key (in options blocks)

**Soft keywords** (only special after `nexus`):
- `service` - Nexus service (in top-level definition or worker reference)
- `endpoint` - Nexus endpoint (in namespace block)

**Configuration:**
- `options` - Options block for activity/workflow/nexus calls

### Symbols

- `->` - Output binding (result assignment)
- `<-` - Promise binding (async declaration)
- `.` - Member access / nexus service.operation separator
- `:` - Block start
- `#` - Comment

### Identifiers

```
IDENT ::= [a-zA-Z_][a-zA-Z0-9_]*
```

Identifiers start with a letter or underscore, followed by any combination of letters, digits, or underscores.

### Literals

```
NUMBER ::= [0-9]+ ['.' [0-9]+]
DURATION ::= NUMBER ('ms' | 's' | 'm' | 'h' | 'd')
STRING ::= '"' [^"]* '"'
```

`NUMBER` and `DURATION` tokens are recognized everywhere. In raw expressions, digits that start a line or follow operators are consumed by the raw text scanner.

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

- `promise` - Non-blocking async operations
- `condition` - Named boolean awaitables
- `set`, `unset` - Condition mutation
- `state` - Workflow state block
- `detach`, `nexus` - Workflow/nexus calls
- `sync`, `async` - Nexus operation types
- `workflow` - Child workflow calls
- `timer` - Durable sleep (with `await`)
- `signal`, `query`, `update` - Handler declarations and await targets
- `await` - Async operation waiting
- `close` - Workflow termination (includes `complete`, `fail`, `continue_as_new`)

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
   - Build condition map from `state:` block declarations
   - Build promise set from `promise` statements in the workflow body
   - Resolve activity calls to activity definitions
   - Resolve workflow calls to workflow definitions
   - Resolve await targets to signal/update/activity/workflow/promise/condition declarations
   - Resolve `set`/`unset` targets to condition declarations
   - Walk signal/query/update handler bodies and resolve references
3. **Report errors:** Undefined references, duplicate definitions, etc.

### Error Handling

The parser and resolver collect multiple errors before failing, allowing users to fix multiple issues in one pass.

Common error types:
- Undefined activity/workflow/signal/update/condition/promise
- Duplicate definitions
- Temporal keywords in activity context
- Invalid await targets (e.g., awaiting a query)
- Condition with result binding (conditions cannot have `-> result`)
- `set`/`unset` on undefined condition
- Worker references undefined workflow, activity, or nexus service
- Duplicate worker, namespace, or nexus service definitions
- Duplicate nexus endpoint names across namespaces
- Namespace references undefined worker
- Worker instantiation missing `task_queue` option
- Nexus endpoint instantiation missing `task_queue` option
- Workers on same task queue with different type sets
- Undefined nexus endpoint (when endpoints exist locally)
- Undefined nexus service (when services exist locally)
- Nexus service has no matching operation
- Detach nexus call with result binding
- Async nexus operation references undefined workflow
- Nexus endpoint routes to task queue with no worker registering the service
- Explicit `task_queue` routing: target activity/workflow not on any worker polling that queue
- Implicit task queue routing: called activity/workflow not on any worker polling the calling workflow's task queue
- Workflow/activity not registered on any instantiated worker (warning)
- Nexus service not referenced by any worker (warning)
- Worker not instantiated in any namespace (warning)
- Empty worker with no registrations (warning)
- Empty namespace with no worker or endpoint instantiations (warning)
- Empty workflow body (warning)
- Empty activity body (warning)
- Unresolved nexus endpoint when no endpoints defined (warning, may be external)
- Unresolved nexus service when no services defined (warning, may be external)
- Unknown option key in `options:` block
- Wrong value type for option key (e.g., number where duration expected)
- Invalid enum value for option key

## Examples

See the `topics/` directory for complete working examples of all language features.

## Grammar Summary

```
file ::= definition*
definition ::= workflow_def | activity_def | worker_def | namespace_def | nexus_service_def

workflow_def ::= 'workflow' IDENT params ['->' return_type] ':'
                 NEWLINE INDENT
                 [state_block]
                 [signal_decl*] [query_decl*] [update_decl*]
                 statement*
                 DEDENT

activity_def ::= 'activity' IDENT params ['->' return_type] ':'
                 NEWLINE INDENT statement* DEDENT

worker_def ::= 'worker' IDENT ':' NEWLINE
               INDENT worker_entry* DEDENT
worker_entry ::= 'workflow' IDENT NEWLINE
               | 'activity' IDENT NEWLINE
               | 'nexus' 'service' IDENT NEWLINE

namespace_def ::= 'namespace' IDENT ':' NEWLINE
                  INDENT namespace_entry* DEDENT
namespace_entry ::= 'worker' IDENT NEWLINE [options_line]
                  | 'nexus' 'endpoint' IDENT NEWLINE [options_line]

nexus_service_def ::= 'nexus' 'service' IDENT ':' NEWLINE
                      INDENT nexus_operation* DEDENT
nexus_operation ::= 'async' IDENT 'workflow' IDENT NEWLINE
                  | 'sync' IDENT params '->' return_type ':' NEWLINE
                    INDENT statement* DEDENT

nexus_call ::= ['detach'] 'nexus' IDENT IDENT '.' IDENT args ['->' result] [NEWLINE options_line]

options_block ::= 'options' ':' NEWLINE INDENT option_entry+ DEDENT
option_entry  ::= IDENT ':' value NEWLINE
                | IDENT ':' NEWLINE INDENT option_entry+ DEDENT
value ::= STRING | DURATION | NUMBER | IDENT
DURATION ::= NUMBER ('ms' | 's' | 'm' | 'h' | 'd')

state_block ::= 'state' ':' NEWLINE INDENT state_stmt* DEDENT
state_stmt ::= condition_decl | raw_stmt
condition_decl ::= 'condition' IDENT NEWLINE

signal_decl ::= 'signal' IDENT params ':' NEWLINE INDENT statement* DEDENT
query_decl ::= 'query' IDENT params '->' return_type ':' NEWLINE INDENT statement* DEDENT
update_decl ::= 'update' IDENT params '->' return_type ':' NEWLINE INDENT statement* DEDENT

statement ::= activity_call | workflow_call | nexus_call | promise_stmt | set_stmt | unset_stmt
            | await_stmt | await_all_block | await_one_block | switch_block
            | if_stmt | for_stmt | close_stmt | return_stmt
            | break_stmt | continue_stmt | assignment

promise_stmt ::= 'promise' IDENT '<-' async_target NEWLINE
set_stmt ::= 'set' IDENT NEWLINE
unset_stmt ::= 'unset' IDENT NEWLINE

await_stmt ::= 'await' (timer_target | signal_target | update_target | activity_target | workflow_target | nexus_target | ident_target) NEWLINE
nexus_target ::= ['detach'] 'nexus' IDENT IDENT '.' IDENT args ['->' result]
ident_target ::= IDENT ['->' result]

await_one_case ::= signal_case | update_case | timer_case | activity_case | workflow_case | nexus_case | await_all_case | ident_case

signal_case ::= 'signal' IDENT ['->' params] ':' NEWLINE [INDENT statement+ DEDENT]

update_case ::= 'update' IDENT ['->' params] ':' NEWLINE [INDENT statement+ DEDENT]

timer_case ::= 'timer' '(' duration ')' ':' NEWLINE [INDENT statement+ DEDENT]

activity_case ::= 'activity' IDENT args ['->' result] ':' NEWLINE [INDENT statement+ DEDENT]

workflow_case ::= ['detach'] 'workflow' IDENT args ['->' result] ':' NEWLINE [INDENT statement+ DEDENT]

nexus_case ::= ['detach'] 'nexus' IDENT IDENT '.' IDENT args ['->' result] ':' NEWLINE [INDENT statement+ DEDENT]

ident_case ::= IDENT ['->' result] ':' NEWLINE [INDENT statement+ DEDENT]

close_stmt ::= 'close' ('complete' | 'fail' | 'continue_as_new') ['(' args ')'] NEWLINE
```
