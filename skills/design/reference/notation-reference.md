# TWF Notation Reference

| Syntax | Meaning |
|--------|---------|
| `activity Name(args) -> result` | Call activity, bind result |
| `workflow Name(args) -> result` | Call child workflow, bind result |
| `nexus Endpoint Service.Op(args) -> result` | Nexus service operation call |
| `detach nexus Endpoint Service.Op(args)` | Fire-and-forget nexus call |
| `promise p <- nexus Endpoint Service.Op(args)` | Start async nexus call |
| `promise p <- activity Name(args)` | Start async activity |
| `promise p <- workflow Name(args)` | Start async child workflow |
| `promise p <- timer(duration)` | Start async timer |
| `promise p <- signal Name` | Promise for signal |
| `await p -> result` | Await promise, bind result |
| `state:` | Workflow state block |
| `condition name` | Named condition (in `state:` block) |
| `set name` | Set condition true |
| `unset name` | Set condition false |
| `await name` | Await condition |
| `detach workflow Name(args)` | Fire-and-forget child workflow |
| `await timer(duration)` | Durable sleep |
| `await signal Name` | Wait for signal |
| `await update Name` | Wait for update |
| `await nexus Endpoint Service.Op(args) -> result` | Wait for nexus call |
| `await one:` | Race: first to complete wins |
| `await all:` | Join: wait for all |
| `options: key: value` | Options block for activity/workflow/nexus calls |
| `-> (Type)` | Return type (always parenthesized) |
| `-> result` | Bind preceding result |
| `close complete\|fail\|continue_as_new(Value)` | End workflow with result, failure, or continuation |
| `if (expr):` / `else:` | Conditional |
| `for (x in collection):` | Bounded loop |
| `for:` | Infinite loop (needs `close continue_as_new` or `close complete`) |
| `switch (expr):` / `case val:` | Multi-branch conditional |
| `close continue_as_new(args)` | Reset history and continue |
| `signal Name(params):` | Signal handler (in workflow, before body) |
| `query Name(params) -> (Type):` | Query handler (in workflow, before body) |
| `update Name(params) -> (Type):` | Update handler (in workflow, before body) |
| `nexus service Name:` | Nexus service definition (top-level) |
| `async OpName workflow WorkflowName` | Async nexus operation (in service body) |
| `sync OpName(params) -> (Type):` | Sync nexus operation (in service body) |
| `worker name:` | Worker type set definition |
| `nexus service Name` (in worker) | Register nexus service on worker |
| `namespace name:` | Namespace definition (deployment with options) |
| `nexus endpoint Name` (in namespace) | Nexus endpoint instantiation with task_queue |

Full grammar: [`LANGUAGE.md`](../../tools/lsp/LANGUAGE.md).
