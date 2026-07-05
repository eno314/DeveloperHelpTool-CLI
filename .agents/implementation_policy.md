# Implementation Policy

This document defines the code design and implementation guidelines for the `DeveloperHelpTool-CLI` project.
Everyone who writes code — human or AI agent — is expected to follow these guidelines.

---

## 1. Library and Framework Usage

**Principle: Maximize use of the standard library. Keep external dependencies to an absolute minimum.**

A third-party library may be adopted only if it satisfies **all four** of the following criteria:

| Criterion | Description |
|-----------|-------------|
| Minimal surface area | Focused on a single responsibility; brings no unwanted functionality |
| Proven longevity | Has a stable track record (guideline: 2+ years since initial release) |
| Actively maintained | Has commits or fixes within the past year |
| High cost to self-implement | Implementing an equivalent in-house is genuinely impractical |

If even one criterion is not met, build it yourself instead.

**Currently approved external dependencies:**

- `github.com/google/go-cmp` — Text diff rendering. Maintained by the Go core team. Implementing a diff algorithm in-house carries a high cost.

**Recording decisions:** When adding a new external library, document in the commit message which of the four criteria it satisfies and why.

---

## 2. Comments

**Principle: Write self-explanatory code. Comments exist solely to explain *why*, not *what* or *how*.**

### Acceptable comments

**Non-obvious implementation rationale** — explain the intent behind code that looks wrong at first glance:

```go
// Amidakuji rules forbid two adjacent horizontal lines from meeting at the same node.
// Skip the next index after placing a horizontal line to enforce this invariant.
l++
```

**Marker flags** — flag locations that require future attention:

```go
// TODO: make the timeout configurable via injection
client := &http.Client{}
```

### Unacceptable comments

- Paraphrasing what the code already says: `// increment i`
- Explaining anything made obvious by the function or variable name
- Decorative block separators: `// --- render ---`

---

## 3. Data-Driven Design

**Principle: Express behavior as data; let code interpret that data.**

### Command routing

Represent command dispatch as a map, not a `switch`. Adding a new command requires only a new map entry.

```go
// Bad: the switch grows with every new command
switch command {
case "amidakuji":
    return features.RunAmidakuji(...)
case "httpdiff":
    return features.RunHttpDiff(...)
}

// Good: data (the map) drives the code
type commandFunc func(args []string, out, err io.Writer) int

var commands = map[string]commandFunc{
    "amidakuji": features.RunAmidakuji,
    "httpdiff":  features.RunHttpDiff,
}

fn, ok := commands[command]
if !ok {
    fmt.Fprintf(errStream, "Unknown command: %s\n", command)
    return 1
}
return fn(args[2:], outStream, errStream)
```

### Table-driven tests

Declare inputs and expected values in a table (slice of structs) and drive them with `t.Run` in a loop.

```go
tests := []struct {
    name         string
    args         []string
    expectedCode int
    expectedOut  string
    expectedErr  string
}{
    {
        name:         "success",
        args:         []string{"--participants", "A,B,C", "--goals", "X,Y,Z"},
        expectedCode: 0,
        expectedOut:  "Results:",
    },
    {
        name:         "no participants",
        args:         []string{},
        expectedCode: 1,
        expectedErr:  "required",
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // ...
    })
}
```

### State management: lookup-table state machine

Model multi-state logic with a transition table (map) rather than a chain of `switch` statements.

```go
type State int
type Event int

const (
    StateIdle    State = iota
    StateRunning
    StateDone
    StateError
)

const (
    EventStart    Event = iota
    EventComplete
    EventFail
)

// The transition table drives the logic.
var transitions = map[State]map[Event]State{
    StateIdle:    {EventStart: StateRunning},
    StateRunning: {EventComplete: StateDone, EventFail: StateError},
}

func transition(current State, event Event) (State, bool) {
    next, ok := transitions[current][event]
    return next, ok
}
```

---

## 4. Pure Functions and Localizing Side Effects

**Principle: Implement logic as pure functions. Confine side effects (I/O, randomness, time, HTTP) to a thin layer near the entry point and inject them as arguments.**

### Side effects and how to handle them

| Side effect | Approach |
|-------------|----------|
| Randomness | Inject `*rand.Rand` as an argument |
| Time | Inject a base `time.Time` or a `func() time.Time` as an argument |
| HTTP | Inject `*http.Client` as an argument |
| File I/O | Inject `io.Writer` / `io.Reader` as arguments |
| Standard output | Inject `io.Writer` as an argument (follows the existing pattern) |

### Example

```go
// Bad: randomness is created inside the logic (untestable, non-deterministic)
func generateBoard(n int) [][]bool {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    // ...
}

// Good: the random source is received as an argument (testable, nearly pure)
func generateBoard(n int, r *rand.Rand) [][]bool {
    // ...
}

// Side effects are created only at the entry point.
func RunAmidakuji(args []string, out, err io.Writer) int {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    board := generateBoard(n, r)
    // ...
}
```

---

## 5. Error Handling

**Principle: Treat errors as values and propagate them as return values from pure functions.**

- **No `panic` in business logic.** Reserve `panic` for genuine programming errors (unreachable code, violated invariants).
- **Wrap errors with context.** Use `fmt.Errorf("...: %w", err)` to attach context at each call site so the caller can diagnose the failure.
  ```go
  u, err := url.JoinPath(host, p)
  if err != nil {
      return fmt.Errorf("joining path for %s and %s: %w", host, p, err)
  }
  ```
- **Classify errors.** Distinguish user-caused errors (invalid arguments) from system errors (network failures) and keep the mapping to exit codes (`0` or `1`) simple.
- **Return validation failures as errors.** Do not print error messages or handle exit codes inside helper or validation functions. Instead, return a standard Go `error` to the caller (e.g., the main entry point), letting the caller decide how to format the output and dictate the program flow.

---

## 6. Minimal Interface Design

**Principle: Define interfaces that demand only what the caller actually needs.**

- **Prefer standard interfaces.** Use `io.Writer`, `io.Reader`, and other standard library interfaces wherever they fit; avoid defining custom ones unnecessarily.
- **Keep custom interfaces small.** When you do define one, limit it to 1–2 methods. Large interfaces are a burden to implementors and make test doubles unnecessarily complex.
- **No mock frameworks.** Do not introduce `golang/mock` or similar libraries for tests. Use `net/http/httptest` for HTTP testing; for other interfaces, write simple inline struct implementations or swap in function pointers.

---

## 7. Function Organization and Single Responsibility

**Principle: Organize files top-down and ensure each function has a single, clear purpose.**

- **Top-Down Ordering:** Place the most important, public functions (like CLI entry points, e.g., `RunHttpDiff`) at the top of the file. Place the private helper functions they call directly below them. This allows readers to understand the high-level flow before diving into the implementation details.
- **Single Responsibility & Naming:** A function should do exactly one thing. If you find yourself adding "And" to a function name (e.g., `parseAndValidateFlags`), it is a strong signal that the function violates the Single Responsibility Principle. Split such functions into distinct operations (e.g., flag parsing as one step, and configuration validation/building as another).

---

## 8. Function Size and Control Flow

**Principle: Keep functions short, flat, and focused. Apply the core concepts of "Five Lines of Code" modified for Go.**

### Function length limit (10-line rule)

A function body should not exceed **10 lines** (excluding the opening and closing braces).
- **Exception:** Standard Go error handling blocks (`if err != nil { return ... }`) are excluded from the line count. The goal is to keep the actual business logic under **5 lines**.

```go
// Good: Clear, flat, and fits within the line limit (excluding error check)
func (s *Service) FetchData(id string) (*Data, error) {
    raw, err := s.client.Get(id)
    if err != nil {
        return nil, fmt.Errorf("getting raw data: %w", err) // Excluded from line count
    }
    
    data, err := parseData(raw)
    if err != nil {
        return nil, fmt.Errorf("parsing: %w", err) // Excluded from line count
    }
    
    return data, nil
}
```

### No `else` and prefer guard clauses

Do not use the `else` keyword. Handle negative or error cases first, return early (Guard Clauses), and keep the happy path at the minimal indentation level.

```go
// Bad: Use of else increases cognitive load and indentation
func checkStatus(status string) string {
    if status == "active" {
        return "OK"
    } else {
        return "Disabled"
    }
}

// Good: No else, flat structure
func checkStatus(status string) string {
    if status == "active" {
        return "OK"
    }
    return "Disabled"
}
```

### Limit nesting levels

Nesting (e.g., nested `if`, `for`, `switch`) inside a single function must be limited to **1 level**. If you need a second level of nesting, extract the inner block into a separate helper function.

```go
// Bad: 2 levels of nesting (for -> if)
func printActive(users []User) {
    for _, u := range users {
        if u.IsActive {
            fmt.Println(u.Name)
        }
    }
}

// Good: Extracting inner logic to keep nesting to 1 level
func printActive(users []User) {
    for _, u := range users {
        printIfActive(u)
    }
}

func printIfActive(u User) {
    if u.IsActive {
        fmt.Println(u.Name)
    }
}
```

---

## 9. Code Review and Quality Gates

**Principle: Every implementation must be reviewed using the `code-review-and-quality` skill.**

Before merging any change or finalizing an implementation, you must explicitly invoke the `code-review-and-quality` skill to evaluate the code against the five axes (Correctness, Readability, Architecture, Security, Performance). Fix any required or critical issues identified during this review before proceeding.

---

## Priority Order

When guidelines conflict, apply them in this order:

1. **Correct** — free of bugs
2. **Readable** — the next reader understands the intent (single responsibility, top-down order)
3. **Testable** — pure functions, localized side effects
4. **Malleable** — data-driven, dependency injection
5. **Minimal dependencies** — fewest external libraries
