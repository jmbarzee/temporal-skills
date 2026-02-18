# TWF Notation Reference

| Syntax | Meaning |
|--------|---------|
| `activity Name(args) -> result` | Call activity, bind result |
| `workflow Name(args) -> result` | Call child workflow, bind result |
| `nexus "namespace" workflow Name(args) -> result` | Cross-namespace workflow call |
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
| `detach nexus "ns" workflow Name(args)` | Fire-and-forget nexus call |
| `await timer(duration)` | Durable sleep |
| `await signal Name` | Wait for signal |
| `await update Name` | Wait for update |
| `await one:` | Race: first to complete wins |
| `await all:` | Join: wait for all |
| `options: key: value` | Options block for activity/workflow calls |
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
| `worker name:` | Worker definition (task queue + registered types) |
| `namespace name` | Worker namespace declaration (inside worker block) |
| `task_queue name` | Worker task queue declaration (inside worker block) |

Full grammar: [`LANGUAGE.md`](../lsp/LANGUAGE.md).
