# TWF Language Spec Changes Summary

This document summarizes recent changes to the TWF language specification and parser implementation.

## Parser Bug Fixes

### 1. Comments Between Signal/Query/Update Declarations
**Fixed:** Comments placed between signal/query/update declarations no longer break parsing.

**Before:** Parser would exit handler declaration section when encountering a comment.
**After:** Parser correctly skips comments and continues parsing handlers.

**Technical:** Renamed `skipNewlines()` to `skipBlankLinesAndComments()` to handle COMMENT tokens in addition to NEWLINE tokens.

---

## Syntax Clarifications

### 2. Return Types Always Parenthesized
**Rule:** Return types must always be enclosed in parentheses.

**Correct:**
```twf
workflow ProcessOrder(order: Order) -> (OrderResult):
query GetStatus() -> (Status):
activity FetchData(id: string) -> (Data):
```

**Incorrect:**
```twf
workflow ProcessOrder(order: Order) -> OrderResult:  # Missing parens
```

**Note:** Return types are optional for workflows and activities. When omitted, there is no arrow at all:
```twf
workflow BackgroundTask(data: Data):  # No return type
activity LogEvent(message: string):   # No return type
```

### 3. No void Convention
**Rule:** For workflows/activities that don't return values, omit the return type entirely. Do not use `-> (void)`.

**Before:**
```twf
activity SendNotification(user: User) -> (void):  # Old style
```

**After:**
```twf
activity SendNotification(user: User):  # Correct style
```

### 4. Options Blocks Require Arrow Syntax
**Rule:** Activity and workflow calls with `options()` blocks MUST use arrow syntax for the result, not assignment syntax.

**Correct:**
```twf
activity ValidateInput(input) -> validated
    options(startToCloseTimeout: 30s)
```

**Incorrect:**
```twf
validated = activity ValidateInput(input)  # Won't work with options block
    options(startToCloseTimeout: 30s)
```

**Note:** Options blocks must be indented on the line following the call.

---

## New Language Features

### 5. heartbeat() Primitive (Activity-Only)
**Added:** `heartbeat()` is now a documented primitive statement available only in activity bodies.

**Syntax:**
```twf
heartbeat_stmt ::= 'heartbeat' '(' [arg_list] ')' NEWLINE
```

**Example:**
```twf
activity ProcessLargeFile(fileId: string) -> (Result):
    file = download(fileId)
    for chunk in file.chunks:
        process(chunk)
        heartbeat(progress: {current: chunk.id, total: file.total})
    return Result{success: true}
```

**Purpose:** Reports progress to Temporal service, enabling activity resumption after failures.

---

## Handler Declarations (Signal/Query/Update)

### 6. Handler Declaration Rules
**Important:** Signal, query, and update handlers have specific declaration requirements:

1. **Declared Once:** Each signal/query/update can only be declared once per workflow
2. **Declaration Location:** Handlers must appear at the beginning of workflows, before any workflow body statements
3. **Referenced Multiple Times:** Use `hint` statements to reference signals/updates multiple times throughout the workflow

**Example:**
```twf
workflow OrderWorkflow(orderId: string) -> (OrderResult):
    # Declare signal ONCE at workflow start
    signal PaymentReceived(transactionId: string, amount: decimal):
        status = "processing"
        activity FulfillOrder(order)

    # Declare query ONCE
    query GetStatus() -> (Status):
        return Status{status: status, orderId: orderId}

    # Workflow body starts here
    status = "awaiting_payment"

    # Reference the signal via hint (can appear multiple times)
    hint signal PaymentReceived
    select:
        timer 24h:
            status = "timeout"

    return OrderResult{status: status}
```

### 7. Query and Update Return Types Required
**Rule:** Queries and updates MUST have return types (always parenthesized).

**Correct:**
```twf
query GetProgress() -> (Progress):
update SetPriority(level: int) -> (int):
```

**Incorrect:**
```twf
query GetProgress():  # Missing return type - ERROR
```

**Rationale:** Queries and updates are always called by external clients who expect responses.

---

## hint Statement

### 8. hint Statement Usage
**Purpose:** Annotates where signals/updates may arrive in the workflow, even when not explicitly awaiting.

**Syntax:**
```twf
hint_stmt ::= 'hint' ('signal' | 'update') IDENT NEWLINE
```

**Key Points:**
- `hint` does NOT define a signal/update - it references one already declared in the workflow
- Each signal/update is declared once at the top, hinted as many times as needed
- Hints tell the execution engine where handler bodies may execute

**Example:**
```twf
workflow ProcessWithInterruption(data: Data) -> (Result):
    signal Pause():
        paused = true

    paused = false

    for item in data.items:
        hint signal Pause  # Handler may execute here
        if (paused):
            break
        activity ProcessItem(item)

    return Result{processed: true}
```

---

## Comments

### 9. Comment Syntax
**Syntax:** Comments start with `#` and continue to end of line.

```twf
# This is a comment
activity DoWork(data: Data):
    # Comments can appear anywhere
    process(data)  # Including after code
```

**Note:** Comments are captured in the AST but do not affect execution semantics.

---

## Versioning Examples Updated

### 10. Removed patched() Function
**Change:** Examples no longer use `patched()` function for version gating.

**Before:**
```twf
hasFraudCheck = patched("add-fraud-check")  # patched() not a language primitive
if (hasFraudCheck):
    activity FraudCheck(order)
```

**After:**
```twf
# Use workflow parameters for version flags
workflow OrderWorkflow(order: Order, enableFraudCheck: bool) -> (OrderResult):
    if (enableFraudCheck):
        activity FraudCheck(order)
```

**Rationale:** `patched()` is not a Temporal primitive. Version flags should be passed as workflow parameters or loaded from external configuration.

---

## Summary for Visualization

### Critical Features for Visualizer:

1. **Handler Declarations:**
   - Signals, queries, and updates are declared at workflow start
   - Each can only be declared ONCE per workflow
   - Use `hint` to reference them later

2. **Return Type Syntax:**
   - Always parenthesized when present: `-> (Type)`
   - Optional for workflows/activities (omit arrow entirely if no return)
   - Required for queries and updates

3. **Options Blocks:**
   - Must be indented on line after call
   - Require arrow syntax: `activity Foo() -> result`

4. **Activity-Only Features:**
   - `heartbeat()` primitive only available in activities

5. **Comments:**
   - Use `#` for comments
   - Can appear between declarations and statements

6. **Constructor Syntax:**
   - Examples use both positional (`Result{x, y}`) and named (`Result{x: x, y: y}`) syntax
   - Spec only documents named syntax (intentional - positional works via opaque expression parsing)

### Visualization Flow Example:

```twf
workflow Example(input: Input) -> (Result):
    # 1. Handler declarations section (visualize as workflow-level definitions)
    signal Pause():
        paused = true
    query GetStatus() -> (Status):
        return Status{paused: paused}

    # 2. Workflow body (visualize as execution flow)
    paused = false
    activity DoWork(input)

    # 3. Hint annotation (visualize as "signal may arrive here")
    hint signal Pause

    activity DoMoreWork(input)
    return Result{success: true}
```

**Key Insight:** Signal/query/update declarations define the "API" of the workflow (what external callers can invoke). The workflow body defines the execution logic. `hint` statements show where that external API may interact with the execution.
