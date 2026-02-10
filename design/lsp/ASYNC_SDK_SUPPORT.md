I’m sorry, but the knowledge sources don’t contain enough information to give a *fully featured* and *complete* cross-SDK review of all async primitives in every Temporal SDK the way you’re asking (covering “all Temporal primitives that can be waited on” per SDK). They only show specific patterns and APIs in a few SDKs (mainly Java, Go, TypeScript, Python, PHP, and Ruby), and even for those, the coverage is partial.

What I can do is summarize and exemplify the async patterns that *are* explicitly shown in the sources, organized by SDK and by your three requested cases:

- Single wait on an external condition / timer
- “Select” / wait-on-many pattern
- Parallel async operations, then wait for all

I’ll keep each snippet self-contained so it’s easy to copy/paste.

---

## Java SDK

### 1. Single wait on an async operation (Activity via `Promise`)

The Java docs show using `Async.function` to invoke an Activity asynchronously and then `get()` to block for the result. [[Java Promise](https://www.javadoc.io/static/io.temporal/temporal-sdk/1.22.3/io/temporal/workflow/package-summary.html)]

```java
import io.temporal.workflow.Async;
import io.temporal.workflow.Promise;

public class MyWorkflowImpl implements MyWorkflow {

  private final MyActivities activities =
      Workflow.newActivityStub(MyActivities.class, options);

  @Override
  public String run(String sourceBucket, String sourceFile) {
    // Asynchronous Activity invocation
    Promise<String> localNamePromise =
        Async.function(activities::download, sourceBucket, sourceFile);

    // Single wait on async result
    String localName = localNamePromise.get(); // blocks until Activity completes
    return localName;
  }
}
```

The sources don’t show a Java timer example, only Activity-based async via `Promise`.

---

### 2. “Select” wait on multiple async operations

The sources mention `Promise.thenApply`, `Promise.handle`, and `Promise.allOf` but don’t show a full “select” example that waits on *different* primitives and reacts to whichever completes first. [[Java Promise](https://www.javadoc.io/static/io.temporal/temporal-sdk/1.22.3/io/temporal/workflow/package-summary.html)]

So I can’t provide a concrete, source-backed “select” snippet for Java that covers “all Temporal primitives that can be waited on”.

---

### 3. Parallel async operations, then wait for all

The Java docs show parallel Activities using `Async.function` and `Promise.allOf`. [[Java Promise](https://www.javadoc.io/static/io.temporal/temporal-sdk/1.22.3/io/temporal/workflow/package-summary.html)]

```java
import io.temporal.workflow.Async;
import io.temporal.workflow.Promise;

public class FileProcessingWorkflowImpl implements FileProcessingWorkflow {

  private final FileActivities activities =
      Workflow.newActivityStub(FileActivities.class, options);

  @Override
  public void processFile(Arguments args) {
    // Start downloads in parallel
    List<Promise<String>> localNamePromises = new ArrayList<>();
    for (String sourceFilename : args.getSourceFilenames()) {
      Promise<String> localNamePromise =
          Async.function(activities::download, args.getSourceBucketName(), sourceFilename);
      localNamePromises.add(localNamePromise);
    }

    // Wait for all to complete
    Promise<List<String>> localNamesPromise = Promise.allOf(localNamePromises);
    List<String> localNames = localNamesPromise.get(); // blocks until all done

    // Use results
    List<String> processedNames = activities.processFiles(localNames);
  }
}
```

---

## Go SDK

### 1. Single wait on an async operation (Activity via `Future`)

The Go docs show `workflow.ExecuteActivity` returning a `Future`, and then `Get()` to block. [[Go ExecuteActivity](https://pkg.go.dev/go.temporal.io/sdk/workflow#hdr-Execute_Activity)]

```go
import (
	"go.temporal.io/sdk/workflow"
)

func SingleWaitWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		// timeouts, etc.
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, MyActivity, "input")

	var result string
	if err := fut.Get(ctx, &result); err != nil {
		return err
	}

	// result is ready here
	return nil
}
```

The sources don’t show a Go timer-only example in this set, but they do show timers used with selectors (below).

---

### 2. “Select” wait on multiple primitives (Future + Channel + Timer)

The Go selector docs show waiting on:

- Activity `Future`
- Signal `ReceiveChannel`
- Timer `Future` (via `workflow.NewTimer`) [[Go Selectors](https://docs.temporal.io/develop/go/selectors)]

```go
import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func SelectWorkflow(ctx workflow.Context) error {
	selector := workflow.NewSelector(ctx)

	// Activity Future
	actFuture := workflow.ExecuteActivity(ctx, ExampleActivity)
	selector.AddFuture(actFuture, func(f workflow.Future) {
		// handle Activity completion
	})

	// Signal channel
	var signalVal string
	signalChan := workflow.GetSignalChannel(ctx, "my-signal")
	selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &signalVal)
		// handle signal
	})

	// Timer Future
	timer := workflow.NewTimer(ctx, 10*time.Second)
	selector.AddFuture(timer, func(f workflow.Future) {
		// handle timeout
	})

	// Block until one of the above is ready
	selector.Select(ctx)

	return nil
}
```

This is the canonical “select” pattern in Go Workflows. [[Go Selectors](https://docs.temporal.io/develop/go/selectors)]

---

### 3. Parallel async operations, then wait for all

The Go docs show starting multiple Activities, storing their `Future`s, then calling `Get()` later. [[Go ExecuteActivity](https://pkg.go.dev/go.temporal.io/sdk/workflow#hdr-Execute_Activity)]

```go
import "go.temporal.io/sdk/workflow"

func ParallelActivitiesWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{/* ... */})

	var futures []workflow.Future
	for _, input := range []string{"a", "b", "c"} {
		f := workflow.ExecuteActivity(ctx, MyActivity, input)
		futures = append(futures, f)
	}

	// Wait for all to complete
	for _, f := range futures {
		var result string
		if err := f.Get(ctx, &result); err != nil {
			return err
		}
		// use result
	}

	return nil
}
```

---

## TypeScript SDK

### 1. Single wait on an external condition / timer

The timers doc shows `sleep` and `condition` as the core primitives. [[TS Timers](https://docs.temporal.io/develop/typescript/timers#asynchronous-design-patterns)]

**Single timer wait:**

```ts
import { sleep } from '@temporalio/workflow';

export async function singleTimerWorkflow(): Promise<void> {
  await sleep('10 seconds'); // wait on durable timer
  // continue after timer fires
}
```

**Single wait on a condition (e.g., Signal-driven flag):** [[TS Message passing](https://docs.temporal.io/develop/typescript/message-passing#add-wait-conditions-to-block)]

```ts
import * as wf from '@temporalio/workflow';

const approve = wf.defineSignal('approve');

export async function conditionWorkflow(): Promise<string> {
  let approvedForRelease = false;
  let approverName: string | undefined;

  wf.setHandler(approve, (input: { name: string }) => {
    approvedForRelease = true;
    approverName = input.name;
  });

  // Wait until condition becomes true
  await wf.condition(() => approvedForRelease);

  return `Approved by ${approverName}`;
}
```

---

### 2. “Select” wait (racing timers and signals)

The timers doc shows using `Promise.race` with `sleep` and with a `Trigger` (signal-driven). [[TS Timers](https://docs.temporal.io/develop/typescript/timers#asynchronous-design-patterns)]

**Race between work and timer:**

```ts
import { sleep } from '@temporalio/workflow';

async function processOrder(ms: number): Promise<void> {
  // some work
}

export async function processOrderWorkflow(opts: {
  orderProcessingMS: number;
  sendDelayedEmailTimeoutMS: number;
}): Promise<void> {
  let processing = true;
  const processOrderPromise = processOrder(opts.orderProcessingMS).then(() => {
    processing = false;
  });

  await Promise.race([
    processOrderPromise,
    sleep(opts.sendDelayedEmailTimeoutMS),
  ]);

  if (processing) {
    await sendNotificationEmail();
    await processOrderPromise;
  }
}
```

**Race between signal and timer:** [[TS Timers](https://docs.temporal.io/develop/typescript/timers#asynchronous-design-patterns)]

```ts
import { defineSignal, sleep, Trigger, setHandler } from '@temporalio/workflow';

const userInteraction = new Trigger<boolean>();
const completeUserInteraction = defineSignal('completeUserInteraction');

export async function userInteractionWorkflow(userId: string) {
  setHandler(completeUserInteraction, () => userInteraction.resolve(true));

  const userInteracted = await Promise.race([
    userInteraction,      // resolves when signal handler calls resolve
    sleep('30 days'),     // durable timer
  ]);

  if (!userInteracted) {
    await sendReminderEmail(userId);
  }
}
```

This effectively “selects” between a signal and a timer.

---

### 3. Parallel async operations, then wait for all

The TypeScript education content shows parallel Activities and `Promise.all`. [[TS Accessing results](https://github.com/temporalio/edu-102-typescript-content/blob/main/improving-temporal-application-code/accessing-results.md)]

```ts
import * as wf from '@temporalio/workflow';
import { activityA, activityB, activityC } from './activities';

export async function parallelActivitiesWorkflow(): Promise<void> {
  // Start Activities in parallel (non-blocking)
  const promiseA = activityA('inputA');
  const promiseB = activityB('inputB');
  const promiseC = activityC('inputC');

  // Wait for all to complete
  const [resultA, resultB, resultC] = await Promise.all([
    promiseA,
    promiseB,
    promiseC,
  ]);

  // use results
}
```

---

## Python SDK

### 1. Single wait on an external condition

The Python docs show `workflow.wait_condition` to wait on a condition, typically driven by a Signal. [[Python message passing](https://docs.temporal.io/develop/python/message-passing#add-wait-conditions-to-block)]

```python
from typing import Optional
from temporalio import workflow

class ApproveInput(workflow.TypedDict):
    name: str

@workflow.defn
class GreetingWorkflow:
    def __init__(self) -> None:
        self.approved_for_release = False
        self.approver_name: Optional[str] = None

    @workflow.signal
    def approve(self, input: ApproveInput) -> None:
        self.approved_for_release = True
        self.approver_name = input["name"]

    @workflow.run
    async def run(self) -> str:
        # Wait until condition becomes true
        await workflow.wait_condition(lambda: self.approved_for_release)
        return f"Approved by {self.approver_name}"
```

The `wait_condition` API is also documented in the core module. [[Python workflow.wait_condition](https://python.temporal.io/temporalio.workflow.html#wait_condition)]

---

### 2. “Select” wait on multiple async operations

The Python core docs show `workflow.wait`, a deterministic version of `asyncio.wait`, which can wait on multiple Futures/Tasks. [[Python workflow.wait](https://python.temporal.io/temporalio.workflow.html)]

However, the provided sources don’t include a concrete example combining Activities, timers, and signals with `workflow.wait`, so I can’t show a fully backed “select across all primitives” snippet.

---

### 3. Parallel async operations, then wait for all

The sources don’t show a concrete Python example of parallel Activities or tasks with `workflow.wait` or `asyncio.gather` inside a Workflow, so I can’t provide a code sample that’s directly supported by these documents.

---

## PHP SDK

### 1. Single wait on an external condition

The PHP docs show `Workflow::await()` to block until a condition is true, typically in an Update handler. [[PHP message passing](https://docs.temporal.io/develop/php/message-passing#message-handler-patterns)]

```php
use Temporal\Workflow;
use Temporal\Workflow\UpdateMethod;

class MyWorkflow
{
    #[UpdateMethod]
    public function myUpdate(UpdateInput $input)
    {
        // Wait until condition is met
        yield Workflow::await(
            fn() => $this->readyForUpdateToExecute($input),
        );

        // continue after condition is true
    }
}
```

This is a single-condition wait; the docs don’t show timers in this snippet.

---

### 2. “Select” wait

The provided PHP docs only cover `Workflow::await()` on a single condition and don’t show a multi-primitive “select” pattern, so I can’t provide a source-backed example that covers all waitable primitives.

---

### 3. Parallel async operations

The sources don’t show a PHP example of parallel Activities or coroutines with a “wait for all” pattern, so I can’t provide a concrete snippet here.

---

## Ruby SDK

### 1. Single wait on an external condition

The Ruby blog describes `Temporalio::Workflow.wait_condition` as a primitive to wait on a block becoming truthy. [[Ruby async constructs](https://temporal.io/blog/temporal-ruby-crash-proof-fibers#additional-async-constructs)]

```ruby
class MyWorkflow < Temporalio::Workflow
  def run
    @approved = false

    # somewhere else, a signal handler would set @approved = true

    Temporalio::Workflow.wait_condition { @approved }

    # continue after condition is true
  end
end
```

The blog notes that `wait_condition` is used under the hood by other async constructs and must be side-effect free. [[Ruby async constructs](https://temporal.io/blog/temporal-ruby-crash-proof-fibers#additional-async-constructs)]

---

### 2. “Select” wait

The Ruby blog doesn’t show a “select” across multiple primitives; it only describes `wait_condition` and `Future` conceptually, so I can’t provide a concrete, source-backed select example.

---

### 3. Parallel async operations, then wait for all

The Ruby blog shows `Temporalio::Workflow::Future.all_of` to wait for multiple concurrent operations (here, child Workflows). [[Ruby async constructs](https://temporal.io/blog/temporal-ruby-crash-proof-fibers#additional-async-constructs)]

```ruby
Temporalio::Workflow::Future.all_of(
  Temporalio::Workflow::Future.new {
    Temporalio::Workflow.execute_child_workflow(SomeChild1, 'my-param1')
  },
  Temporalio::Workflow::Future.new {
    Temporalio::Workflow.execute_child_workflow(SomeChild2, 'my-param2')
  },
  Temporalio::Workflow::Future.new {
    Temporalio::Workflow.execute_child_workflow(SomeChild3, 'my-param3')
  }
).wait
```

This is a clear “parallel then wait for all” pattern in Ruby Workflows.

---

## What’s missing / uncertain

- The sources don’t cover:
  - Swift SDK async patterns (only mention example repos). [[Swift blog](https://temporal.io/blog/temporal-now-supports-swift#examples-to-help-you-get-started-learning-by-doing)]
  - .NET SDK async patterns.
  - A complete enumeration of “all Temporal primitives that can be waited on” per SDK.
- For Python and PHP, we only see condition-based waits, not explicit parallel-Activity examples.
- For Java, we see `Promise` and `Promise.allOf`, but not a multi-primitive “select” example.

If you’d like, you can narrow the scope (for example: “only Go and TypeScript, and only Activities + timers + signals”), and I can reorganize these into a more tabular or side‑by‑side format for easier export.