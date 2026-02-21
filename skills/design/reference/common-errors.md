# Common Errors

Common `twf check` errors and how to fix them.

## Resolve Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `undefined activity: Foo` | Activity `Foo` is called but not defined | Add `activity Foo(...):` definition to the file |
| `undefined workflow: Foo` | Child workflow `Foo` is called but not defined | Add `workflow Foo(...):` definition to the file |
| `undefined signal: Foo` | `await signal Foo` or `signal Foo:` case but no signal handler declared | Add `signal Foo(...):` declaration inside the workflow, before the body |
| `undefined update: Foo` | `await update Foo` or `update Foo:` case but no update handler declared | Add `update Foo(...) -> (Type):` declaration inside the workflow, before the body |
| `undefined condition: Foo` | `set Foo`, `unset Foo`, or `await Foo` but no condition declared | Add `condition Foo` inside the workflow's `state:` block |
| `undefined promise or condition: Foo` | `await Foo` or `Foo:` case in `await one` but `Foo` is not a promise or condition | Add `promise Foo <- ...` in the workflow body or `condition Foo` in the `state:` block |
| `duplicate workflow definition: Foo` | Two `workflow Foo` definitions in the same file | Remove or rename the duplicate |
| `duplicate activity definition: Foo` | Two `activity Foo` definitions in the same file | Remove or rename the duplicate |
| `condition "Foo" cannot have a result binding` | `await Foo -> result` where `Foo` is a condition | Conditions are boolean — remove the `-> result` binding |

## Parse Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `<keyword> is not allowed in activity body` | Using a temporal primitive (`workflow`, `activity`, `timer`, `signal`, `await`, etc.) inside an activity definition or query handler | Move the temporal primitive to a workflow. Activities can only contain non-temporal logic. |
| `expected ( after return type ->` | Return type not parenthesized: `-> Result` | Use `-> (Result)` — return types must be wrapped in parentheses |
| `expected ( after if` / `expected ( after for` | Missing parentheses around condition/iterator | Use `if (expr):` / `for (x in items):` |
| `unexpected token <tok> at top level` | Statement or keyword that doesn't start a workflow or activity definition | Ensure all top-level items are `workflow`, `activity`, `worker`, `namespace`, or `nexus service` definitions |
| `unexpected token <tok> in await one case` | Invalid case type inside `await one:` block | Cases must be `signal`, `update`, `timer`, `activity`, `workflow`, an identifier, or `await all` |
