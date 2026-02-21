# worker

## DSL

```twf
worker orderTypes:
    workflow ProcessOrder
    activity ValidateOrder
    activity ChargePayment

namespace default:
    worker orderTypes
        options:
            task_queue: "orders"
```

## Go

```go
func main() {
    c, err := client.Dial(client.Options{})
    if err != nil {
        log.Fatalln("Unable to create client", err)
    }
    defer c.Close()

    w := worker.New(c, "orders", worker.Options{})

    w.RegisterWorkflow(ProcessOrder)
    w.RegisterActivity(&Activities{/* dependencies */})

    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Unable to start worker", err)
    }
}
```

## Notes

- `worker.New(client, taskQueue, options)` — task queue comes from namespace `options: task_queue`
- `RegisterWorkflow(func)` — one call per workflow in the worker's type set
- `RegisterActivity(struct)` — register the activity struct (all exported methods become activities) or individual functions
- `worker.InterruptCh()` for graceful shutdown on SIGINT/SIGTERM
- Multiple workers in same namespace — multiple `worker.New` calls with different task queues in the same `main()`
- For nexus services on the same worker, see [nexus-service-def.md](./nexus-service-def.md)
