# Workers, Task Queues, and Deployment Topology

> **Example:** [`task-queues.twf`](./task-queues.twf)

Workers group type registrations. Task queues route work to workers. Namespaces instantiate workers with deployment options. Together they answer: **what runs together, how work reaches it, and where it's deployed.**

---

## Worker Type Sets

Workers are reusable type sets that list which workflows, activities, and nexus services belong together:

```twf
worker orderTypes:
    workflow ProcessOrder
    workflow CancelOrder
    activity ChargePayment
    activity SendNotification
```

Workers contain only type references — no deployment config. Naming: `lowerCamelCase`.

---

## Namespace Instantiation

Namespaces instantiate workers with deployment options (task queue, concurrency limits, etc.):

```twf
namespace orders:
    worker orderTypes
        options:
            task_queue: "orderProcessing"
            max_concurrent_activity_executions: 50
```

The same worker type set can be reused across namespaces:

```twf
namespace staging:
    worker orderTypes
        options:
            task_queue: "staging-orders"
```

### What the Resolver Validates

- **Undefined references** — Catch typos (e.g., referencing a workflow or worker that doesn't exist)
- **Coverage gaps** — Warn when a defined workflow/activity isn't registered on any instantiated worker
- **Task queue coherence** — Error when different workers on the same queue register different type sets
- **Missing configuration** — Error when a worker instantiation is missing the required `task_queue` option

### Rules

- Workers contain only `workflow`, `activity`, and `nexus service` entries (type set only, no deployment config)
- Each worker instantiation in a namespace requires a `task_queue` option
- Worker names use lowerCamelCase; workflow/activity names keep UpperCamelCase
- Multiple workers can be instantiated on the same task queue (but must register the same type sets)
- Workers not instantiated in any namespace produce warnings

---

## Task Queue Design Decisions

### Single vs Multiple Task Queues

| Single Queue | Multiple Queues |
|--------------|-----------------|
| Simple deployment | More operational complexity |
| All workers handle all work | Workers specialize |
| Scaling affects everything | Scale queues independently |
| One failure domain | Isolated failure domains |

### When to Use Separate Task Queues

| Use Case | Rationale |
|----------|-----------|
| **Different resource requirements** | CPU-heavy vs I/O-heavy work |
| **Different scaling characteristics** | Bursty vs steady workloads |
| **Isolation requirements** | Tenant isolation, security boundaries |
| **Priority handling** | High-priority vs batch processing |
| **Geographic distribution** | Region-specific workers |
| **Specialized capabilities** | GPU workers, licensed software |

---

## Task Queue Patterns

### Priority Queues

Separate queues for different priorities:

```twf
workflow OrderWorkflow(order: Order) -> (Result):
    if (order.priority == "express"):
        activity ProcessOrder(order)
            options:
                task_queue: "high-priority"
    else:
        activity ProcessOrder(order)
            options:
                task_queue: "standard"
```

### Tenant Isolation

Separate queues per tenant:

```twf
workflow TenantWorkflow(tenantId: string, data: Data) -> (Result):
    # Route to tenant-specific queue
    activity ProcessData(data)
        options:
            task_queue: "tenant-{tenantId}"
```

### Capability-Based Routing

Route based on required capabilities:

```twf
workflow MediaWorkflow(media: Media) -> (Result):
    if (media.type == "video"):
        # Needs GPU workers
        activity TranscodeVideo(media)
            options:
                task_queue: "gpu-workers"
    else:
        activity ProcessImage(media)
            options:
                task_queue: "standard-workers"
```

### Geographic Routing

Route to region-specific workers:

```twf
workflow GlobalWorkflow(request: Request) -> (Result):
    # Route to nearest region
    activity ProcessLocally(request)
        options:
            task_queue: "workers-{request.region}"
```

---

## Anti-Patterns

### One Queue Per Workflow Type

```twf
# BAD: Unnecessary complexity — one queue per type
# Results in many queues, complex deployment

# GOOD: Shared queue unless isolation needed
# Put related workflows on the same worker and task queue
```

### Dynamic Queue Names Without Cleanup

```twf
# BAD: Creates queue per request (never cleaned up)
workflow Process(requestId: string):
    activity DoWork()
        options:
            task_queue: "request-{requestId}"  # Unbounded queues!

# GOOD: Bounded set of queues
workflow Process(request: Request):
    queue = selectQueue(request.priority)  # "high", "medium", "low"
    activity DoWork()
        options:
            task_queue: queue
```
