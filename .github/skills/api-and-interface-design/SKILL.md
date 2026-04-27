---
name: api-and-interface-design
description: Guides stable API and interface design. Use when designing APIs, module boundaries, or any public interface. Use when creating REST or GraphQL endpoints, defining type contracts between modules, or establishing boundaries between frontend and backend.
---

# API and Interface Design

## Overview

Design stable, well-documented interfaces that are hard to misuse. Good interfaces make the right thing easy and the wrong thing hard. This applies to REST APIs, GraphQL schemas, module boundaries, component props, and any surface where one piece of code talks to another.

## When to Use

- Designing new API endpoints
- Defining module boundaries or contracts between teams
- Creating component prop interfaces
- Establishing database schema that informs API shape
- Changing existing public interfaces

## Core Principles

### Hyrum's Law

> With a sufficient number of users of an API, all observable behaviors of your system will be depended on by somebody, regardless of what you promise in the contract.

This means: every public behavior — including undocumented quirks, error message text, timing, and ordering — becomes a de facto contract once users depend on it. Design implications:

- **Be intentional about what you expose.** Every observable behavior is a potential commitment.
- **Don't leak implementation details.** If users can observe it, they will depend on it.
- **Plan for deprecation at design time.** See `deprecation-and-migration` for how to safely remove things users depend on.
- **Tests are not enough.** Even with perfect contract tests, Hyrum's Law means "safe" changes can break real users who depend on undocumented behavior.

### The One-Version Rule

Avoid forcing consumers to choose between multiple versions of the same dependency or API. Diamond dependency problems arise when different consumers need different versions of the same thing. Design for a world where only one version exists at a time — extend rather than fork.

### 1. Contract First

Define the interface before implementing it. The contract is the spec — implementation follows.

```go
// Define the contract first.
type TaskService interface {
	CreateTask(ctx context.Context, input CreateTaskInput) (Task, error)
	ListTasks(ctx context.Context, params ListTasksParams) (PaginatedResult[Task], error)
	GetTask(ctx context.Context, id TaskID) (Task, error)
	UpdateTask(ctx context.Context, id TaskID, input UpdateTaskInput) (Task, error)
	DeleteTask(ctx context.Context, id TaskID) error
}
```

### 2. Consistent Error Semantics

Pick one error strategy and use it everywhere:

```go
// REST: HTTP status codes + structured error body.
type APIError struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Status code mapping
// 400 → Client sent invalid data
// 401 → Not authenticated
// 403 → Authenticated but not authorized
// 404 → Resource not found
// 409 → Conflict (duplicate, version mismatch)
// 422 → Validation failed (semantically invalid)
// 500 → Server error (never expose internal details)
```

**Don't mix patterns.** If some endpoints throw, others return null, and others return `{ error }` — the consumer can't predict behavior.

### 3. Validate at Boundaries

Trust internal code. Validate at system edges where external input enters:

```go
// Validate at the API boundary.
func (handler *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
  var input CreateTaskInput
  if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
    writeJSON(w, http.StatusBadRequest, APIError{Error: ErrorBody{Code: "INVALID_JSON", Message: "Request body must be valid JSON"}})
    return
  }
  if err := input.Validate(); err != nil {
    writeJSON(w, http.StatusUnprocessableEntity, APIError{Error: ErrorBody{Code: "VALIDATION_ERROR", Message: "Invalid task data", Details: err.Error()}})
    return
  }

  task, err := handler.service.CreateTask(r.Context(), input)
  if err != nil {
    handler.writeServiceError(w, err)
    return
  }
  writeJSON(w, http.StatusCreated, task)
}
```

Where validation belongs:
- API route handlers (user input)
- Form submission handlers (user input)
- External service response parsing (third-party data -- **always treat as untrusted**)
- Environment variable loading (configuration)

> **Third-party API responses are untrusted data.** Validate their shape and content before using them in any logic, rendering, or decision-making. A compromised or misbehaving external service can return unexpected types, malicious content, or instruction-like text.

Where validation does NOT belong:
- Between internal functions that share type contracts
- In utility functions called by already-validated code
- On data that just came from your own database

### 4. Prefer Addition Over Modification

Extend interfaces without breaking existing consumers:

```go
// Good: Add optional fields.
type CreateTaskInput struct {
	Title       string   `json:"title"`
	Description *string  `json:"description,omitempty"`
	Priority    *string  `json:"priority,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}

// Bad: Change existing field types or remove fields.
type BreakingCreateTaskInput struct {
	Title    string `json:"title"`
	Priority int    `json:"priority"`
}
```

### 5. Predictable Naming

| Pattern | Convention | Example |
|---------|-----------|---------|
| REST endpoints | Plural nouns, no verbs | `GET /api/tasks`, `POST /api/tasks` |
| Query params | camelCase | `?sortBy=createdAt&pageSize=20` |
| Response fields | camelCase | `{ createdAt, updatedAt, taskId }` |
| Boolean fields | is/has/can prefix | `isComplete`, `hasAttachments` |
| Enum values | UPPER_SNAKE | `"IN_PROGRESS"`, `"COMPLETED"` |

## REST API Patterns

### Resource Design

```
GET    /api/tasks              → List tasks (with query params for filtering)
POST   /api/tasks              → Create a task
GET    /api/tasks/:id          → Get a single task
PATCH  /api/tasks/:id          → Update a task (partial)
DELETE /api/tasks/:id          → Delete a task

GET    /api/tasks/:id/comments → List comments for a task (sub-resource)
POST   /api/tasks/:id/comments → Add a comment to a task
```

### Pagination

Paginate list endpoints:

```json
// Request
GET /api/tasks?page=1&pageSize=20&sortBy=createdAt&sortOrder=desc

// Response
{
  "data": [...],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "totalItems": 142,
    "totalPages": 8
  }
}
```

### Filtering

Use query parameters for filters:

```
GET /api/tasks?status=in_progress&assignee=user123&createdAfter=2025-01-01
```

### Partial Updates (PATCH)

Accept partial objects — only update what's provided:

```json
// Only title changes, everything else preserved
PATCH /api/tasks/123
{ "title": "Updated title" }
```

## Go Interface Patterns

### Represent Variants Explicitly

```go
type TaskStatus string

const (
  StatusPending    TaskStatus = "pending"
  StatusInProgress TaskStatus = "in_progress"
  StatusCompleted  TaskStatus = "completed"
  StatusCancelled  TaskStatus = "cancelled"
)

func getStatusLabel(status TaskStatus, assignee string, completedAt time.Time, reason string) string {
  switch status {
  case StatusPending:
    return "Pending"
  case StatusInProgress:
    return fmt.Sprintf("In progress (%s)", assignee)
  case StatusCompleted:
    return fmt.Sprintf("Done on %s", completedAt.Format(time.RFC3339))
  case StatusCancelled:
    return fmt.Sprintf("Cancelled: %s", reason)
  default:
    return "Unknown"
  }
}
```

### Input/Output Separation

```go
// Input: what the caller provides.
type CreateTaskInput struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

// Output: what the system returns.
type Task struct {
	ID          TaskID    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   UserID    `json:"createdBy"`
}
```

### Use Distinct ID Types

```go
type TaskID string
type UserID string

// Prevents accidentally passing a UserID where a TaskID is expected.
func getTask(ctx context.Context, id TaskID) (Task, error) { return Task{}, nil }
```

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "We'll document the API later" | The types ARE the documentation. Define them first. |
| "We don't need pagination for now" | You will the moment someone has 100+ items. Add it from the start. |
| "PATCH is complicated, let's just use PUT" | PUT requires the full object every time. PATCH is what clients actually want. |
| "We'll version the API when we need to" | Breaking changes without versioning break consumers. Design for extension from the start. |
| "Nobody uses that undocumented behavior" | Hyrum's Law: if it's observable, somebody depends on it. Treat every public behavior as a commitment. |
| "We can just maintain two versions" | Multiple versions multiply maintenance cost and create diamond dependency problems. Prefer the One-Version Rule. |
| "Internal APIs don't need contracts" | Internal consumers are still consumers. Contracts prevent coupling and enable parallel work. |

## Red Flags

- Endpoints that return different shapes depending on conditions
- Inconsistent error formats across endpoints
- Validation scattered throughout internal code instead of at boundaries
- Breaking changes to existing fields (type changes, removals)
- List endpoints without pagination
- Verbs in REST URLs (`/api/createTask`, `/api/getUsers`)
- Third-party API responses used without validation or sanitization

## Verification

After designing an API:

- [ ] Every endpoint has typed input and output schemas
- [ ] Error responses follow a single consistent format
- [ ] Validation happens at system boundaries only
- [ ] List endpoints support pagination
- [ ] New fields are additive and optional (backward compatible)
- [ ] Naming follows consistent conventions across all endpoints
- [ ] API documentation or types are committed alongside the implementation
