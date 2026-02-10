# Task Queues and Worker Scaling

> **Example:** [`examples/task-queues.twf`](./examples/task-queues.twf)

Task queues route work to workers. Understanding task queue design is essential for scaling, isolation, and performance.

## Task Queue Fundamentals

### How Task Queues Work

```
┌─────────────────────────────────────────────────────────────┐
│                      Temporal Server                        │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │ Task Queue: A   │  │ Task Queue: B   │                  │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │                  │
│  │ │ Task 1      │ │  │ │ Task 4      │ │                  │
│  │ │ Task 2      │ │  │ │ Task 5      │ │                  │
│  │ │ Task 3      │ │  │ └─────────────┘ │                  │
│  │ └─────────────┘ │  └────────┬────────┘                  │
│  └────────┬────────┘           │                           │
└───────────┼────────────────────┼───────────────────────────┘
            │ poll               │ poll
    ┌───────▼───────┐    ┌───────▼───────┐
    │   Worker 1    │    │   Worker 2    │
    │  (Queue A)    │    │  (Queue B)    │
    └───────────────┘    └───────────────┘
```

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Task Queue** | Named queue where tasks wait for workers |
| **Worker** | Process that polls task queue and executes tasks |
| **Workflow Task** | Task to execute workflow code |
| **Activity Task** | Task to execute activity code |
| **Poller** | Thread/goroutine that polls for tasks |

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

## Worker Configuration

### Basic Worker Setup

```
worker = Worker(
    client: temporal_client,
    task_queue: "main-queue",
    workflows: [OrderWorkflow, PaymentWorkflow],
    activities: [ValidateOrder, ProcessPayment]
)

worker.run()
```

### Poller Configuration

```
worker = Worker(
    task_queue: "main-queue",
    
    # Workflow task pollers (usually low, workflows are fast)
    max_concurrent_workflow_task_pollers: 4,
    
    # Activity task pollers (scale based on activity concurrency needs)
    max_concurrent_activity_task_pollers: 10,
    
    # Max concurrent executions
    max_concurrent_workflow_tasks: 100,
    max_concurrent_activities: 50
)
```

### Poller Tuning Guidelines

| Setting | Low Value | High Value | Guidance |
|---------|-----------|------------|----------|
| Workflow pollers | Underutilized CPU | Wasted connections | 2-4 usually sufficient |
| Activity pollers | Underutilized workers | Connection overhead | Match expected concurrency |
| Concurrent workflows | Limited throughput | Memory pressure | Based on workflow complexity |
| Concurrent activities | Limited throughput | Resource exhaustion | Based on activity resource needs |

---

## Scaling Patterns

### Horizontal Scaling

Add more worker instances polling the same queue:

```
# Worker 1, 2, 3... all poll same queue
┌─────────────────┐
│ Task Queue: A   │
└───────┬─────────┘
        │
   ┌────┴────┐
   ▼    ▼    ▼
Worker Worker Worker
  1      2      3
```

**Characteristics:**
- Tasks distributed across workers
- No coordination needed
- Linear scaling (mostly)
- All workers must have same capabilities

### Vertical Scaling

Increase resources per worker:

```
worker = Worker(
    task_queue: "compute-heavy",
    max_concurrent_activities: 100,  # Increased from 50
    # ... on a larger machine
)
```

**Characteristics:**
- Fewer instances to manage
- Limited by single machine
- May have resource contention

### Queue-Based Scaling

Different queues scale independently:

```
┌─────────────────┐     ┌─────────────────┐
│ Queue: fast     │     │ Queue: batch    │
└───────┬─────────┘     └───────┬─────────┘
        │                       │
   ┌────┴────┐            ┌─────┴─────┐
   ▼    ▼    ▼            ▼           ▼
Worker Worker Worker   Worker      Worker
  1      2      3        1            2

Fast: 3 workers (low latency)
Batch: 2 workers (cost efficient)
```

---

## Task Queue Patterns

### Priority Queues

Separate queues for different priorities:

```
workflow OrderWorkflow(order: Order) -> Result:
    if order.priority == "express":
        activity ProcessOrder(order):
            task_queue: "high-priority"
    else:
        activity ProcessOrder(order):
            task_queue: "standard"
```

Worker deployment:
```
# More workers on high-priority queue
high_priority_workers: 10
standard_workers: 5
```

### Tenant Isolation

Separate queues per tenant:

```
workflow TenantWorkflow(tenantId: string, data: Data) -> Result:
    # Route to tenant-specific queue
    activity ProcessData(data):
        task_queue: "tenant-{tenantId}"
```

```
# Deploy workers per tenant
for tenant in tenants:
    Worker(task_queue: "tenant-{tenant.id}").run()
```

### Capability-Based Routing

Route based on required capabilities:

```
workflow MediaWorkflow(media: Media) -> Result:
    if media.type == "video":
        # Needs GPU workers
        activity TranscodeVideo(media):
            task_queue: "gpu-workers"
    else:
        activity ProcessImage(media):
            task_queue: "standard-workers"
```

### Geographic Routing

Route to region-specific workers:

```
workflow GlobalWorkflow(request: Request) -> Result:
    # Route to nearest region
    activity ProcessLocally(request):
        task_queue: "workers-{request.region}"
```

---

## Sticky Execution

Workflows "stick" to workers for cache efficiency.

### How Sticky Execution Works

```
1. Workflow starts on Worker A
2. Worker A caches workflow state
3. Next workflow task routed to Worker A (sticky)
4. If Worker A unavailable, falls back to any worker
```

### Sticky Queue Configuration

```
worker = Worker(
    task_queue: "main",
    
    # How long workflow sticks to this worker
    sticky_schedule_to_start_timeout: 5s,
    
    # Max cached workflows
    max_cached_workflows: 1000
)
```

### When Sticky Execution Matters

| Scenario | Impact |
|----------|--------|
| Complex workflows | Cache miss = full replay |
| High workflow volume | Cache misses = high CPU |
| Worker restarts | All workflows replay |
| Workflow completion | Cache entry freed |

---

## Multi-Queue Workers

Workers can poll multiple queues:

```
worker = Worker(
    client: client,
    task_queue: "primary",
    additional_task_queues: ["secondary", "overflow"],
    workflows: [Workflow1, Workflow2],
    activities: [Activity1, Activity2]
)
```

### Use Cases

| Use Case | Configuration |
|----------|---------------|
| Primary + overflow | Main queue + spike handling |
| Shared + dedicated | Common activities + specialized |
| Migration | Old queue + new queue during transition |

---

## Monitoring and Observability

### Key Metrics

| Metric | What It Tells You |
|--------|-------------------|
| `schedule_to_start_latency` | How long tasks wait in queue |
| `task_queue_backlog` | Tasks waiting for workers |
| `worker_task_slots_available` | Worker capacity |
| `poller_utilization` | Are pollers busy or idle? |

### Scaling Indicators

| Indicator | Action |
|-----------|--------|
| High schedule_to_start latency | Add workers or pollers |
| Growing backlog | Add workers |
| Low poller utilization | Reduce pollers (save connections) |
| High worker CPU | Reduce concurrent activities or add workers |

### Health Checks

```
# Check if workers are polling
temporal task-queue describe --task-queue main-queue

# Check backlog
temporal task-queue get-build-ids --task-queue main-queue
```

---

## Anti-Patterns

### One Queue Per Workflow Type

```
# BAD: Unnecessary complexity
workflow OrderWorkflow():
    task_queue: "order-queue"

workflow PaymentWorkflow():
    task_queue: "payment-queue"

# Results in many queues, complex deployment

# GOOD: Shared queue unless isolation needed
workflow OrderWorkflow():
    task_queue: "main-queue"

workflow PaymentWorkflow():
    task_queue: "main-queue"
```

### Too Many Pollers

```
# BAD: Excessive connections
worker = Worker(
    max_concurrent_workflow_task_pollers: 100,  # Way too many
    max_concurrent_activity_task_pollers: 100
)

# GOOD: Reasonable poller counts
worker = Worker(
    max_concurrent_workflow_task_pollers: 4,
    max_concurrent_activity_task_pollers: 20
)
```

### No Backpressure

```
# BAD: Accept unlimited concurrent activities
worker = Worker(
    max_concurrent_activities: 10000  # Will exhaust resources
)

# GOOD: Match to actual capacity
worker = Worker(
    max_concurrent_activities: 50  # Based on worker resources
)
```

### Dynamic Queue Names Without Cleanup

```
# BAD: Creates queue per request (never cleaned up)
workflow Process(requestId: string):
    activity DoWork():
        task_queue: "request-{requestId}"  # Unbounded queues!

# GOOD: Bounded set of queues
workflow Process(request: Request):
    queue = selectQueue(request.priority)  # "high", "medium", "low"
    activity DoWork():
        task_queue: queue
```

---

## Deployment Considerations

### Worker Deployment Checklist

- [ ] Task queue name matches workflow/activity routing
- [ ] Poller counts appropriate for expected load
- [ ] Concurrent execution limits match resources
- [ ] Health checks configured
- [ ] Graceful shutdown handling
- [ ] Multiple instances for availability

### Rolling Updates

```
# Workers can be updated independently
# Temporal handles routing to available workers

1. Deploy new worker version
2. New workers start polling
3. Old workers finish current tasks
4. Old workers shut down
5. All traffic on new workers
```

### Graceful Shutdown

```
worker = Worker(task_queue: "main")

on_shutdown_signal:
    worker.stop()  # Stop accepting new tasks
    # Finish in-progress tasks
    # Then exit
```
