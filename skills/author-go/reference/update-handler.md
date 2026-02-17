# update handler

## DSL

```twf
update ChangePlan(newPlan: string) -> (ChangeResult):
    activity ValidatePlan(newPlan) -> validation
    if (validation.valid):
        plan = newPlan
        return ChangeResult{success: true, plan: plan}
    else:
        return ChangeResult{success: false, error: validation.reason}
```

## Go

```go
err := workflow.SetUpdateHandlerWithOptions(ctx, "ChangePlan",
    func(ctx workflow.Context, newPlan string) (ChangeResult, error) {
        var validation Validation
        err := workflow.ExecuteActivity(ctx, ValidatePlan, newPlan).Get(ctx, &validation)
        if err != nil {
            return ChangeResult{}, err
        }
        if validation.Valid {
            plan = newPlan
            return ChangeResult{Success: true, Plan: plan}, nil
        }
        return ChangeResult{Success: false, Error: validation.Reason}, nil
    },
    workflow.UpdateHandlerOptions{},
)
if err != nil {
    return Result{}, err
}
```

## Notes

- Update handlers receive `workflow.Context` as first param (unlike queries) — they can call activities and use temporal primitives
- Update handlers can modify workflow state (unlike queries)
- The caller blocks until the handler returns
- Update handlers cannot call `close` — only the main workflow body can terminate the workflow
- Register updates early in the workflow function (before any blocking calls) so they're available immediately
