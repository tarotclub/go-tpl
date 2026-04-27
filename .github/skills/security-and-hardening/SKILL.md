---
name: security-and-hardening
description: Hardens code against vulnerabilities. Use when handling user input, authentication, data storage, or external integrations. Use when building any feature that accepts untrusted data, manages user sessions, or interacts with third-party services.
---

# Security and Hardening

## Overview

Security-first development practices for web applications. Treat every external input as hostile, every secret as sacred, and every authorization check as mandatory. Security isn't a phase — it's a constraint on every line of code that touches user data, authentication, or external systems.

## When to Use

- Building anything that accepts user input
- Implementing authentication or authorization
- Storing or transmitting sensitive data
- Integrating with external APIs or services
- Adding file uploads, webhooks, or callbacks
- Handling payment or PII data

## The Three-Tier Boundary System

### Always Do (No Exceptions)

- **Validate all external input** at the system boundary (API routes, form handlers)
- **Parameterize all database queries** — never concatenate user input into SQL
- **Encode output** to prevent XSS (use framework auto-escaping, don't bypass it)
- **Use HTTPS** for all external communication
- **Hash passwords** with bcrypt/scrypt/argon2 (never store plaintext)
- **Set security headers** (CSP, HSTS, X-Frame-Options, X-Content-Type-Options)
- **Use httpOnly, secure, sameSite cookies** for sessions
- **Run `govulncheck ./...`** (or equivalent) before every release

### Ask First (Requires Human Approval)

- Adding new authentication flows or changing auth logic
- Storing new categories of sensitive data (PII, payment info)
- Adding new external service integrations
- Changing CORS configuration
- Adding file upload handlers
- Modifying rate limiting or throttling
- Granting elevated permissions or roles

### Never Do

- **Never commit secrets** to version control (API keys, passwords, tokens)
- **Never log sensitive data** (passwords, tokens, full credit card numbers)
- **Never trust client-side validation** as a security boundary
- **Never disable security headers** for convenience
- **Never use `eval()` or `innerHTML`** with user-provided data
- **Never store sessions in client-accessible storage** (localStorage for auth tokens)
- **Never expose stack traces** or internal error details to users

## OWASP Top 10 Prevention

### 1. Injection (SQL, NoSQL, OS Command)

```go
// BAD: SQL injection via string concatenation.
query := "SELECT * FROM users WHERE id = '" + userID + "'"

// GOOD: Parameterized query.
row := db.QueryRowContext(ctx, "SELECT id, email FROM users WHERE id = $1", userID)

// GOOD: Prepared statement for repeated calls.
stmt, err := db.PrepareContext(ctx, "SELECT id, email FROM users WHERE id = $1")
```

### 2. Broken Authentication

```go
// Password hashing.
hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
if err != nil {
	return err
}
if err := bcrypt.CompareHashAndPassword(hash, []byte(plaintext)); err != nil {
	return err
}

// Session cookie management.
http.SetCookie(w, &http.Cookie{
	Name:     "session",
	Value:    sessionToken,
	HttpOnly: true,
	Secure:   true,
	SameSite: http.SameSiteLaxMode,
	MaxAge:   24 * 60 * 60,
	Path:     "/",
})
```

### 3. Cross-Site Scripting (XSS)

```go
// BAD: Writing untrusted HTML directly to the response.
fmt.Fprintf(w, userInput)

// GOOD: Use html/template auto-escaping.
tmpl.Execute(w, struct{ Message string }{Message: userInput})

// If you must allow rich HTML, sanitize first with a reviewed sanitizer.
policy := bluemonday.UGCPolicy()
clean := policy.Sanitize(userInput)
```

### 4. Broken Access Control

```go
// Always check authorization, not just authentication.
func (handler *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
  user := auth.UserFromContext(r.Context())
  task, err := handler.service.GetTask(r.Context(), TaskID(chi.URLParam(r, "id")))
  if err != nil {
    handler.writeServiceError(w, err)
    return
  }
  if task.OwnerID != user.ID {
    writeJSON(w, http.StatusForbidden, APIError{Error: ErrorBody{Code: "FORBIDDEN", Message: "Not authorized to modify this task"}})
    return
  }
  // Proceed with update.
}
```

### 5. Security Misconfiguration

```go
// Security headers.
func securityHeaders(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("Content-Security-Policy", "default-src 'self'")
    w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    next.ServeHTTP(w, r)
  })
}

// CORS — restrict to known origins.
handler := cors.Handler(cors.Options{
  AllowedOrigins: []string{"https://app.example.com"},
  AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE"},
  AllowCredentials: true,
})(router)
```

### 6. Sensitive Data Exposure

```go
// Never return sensitive fields in API responses.
func sanitizeUser(user UserRecord) PublicUser {
  return PublicUser{
    ID:    user.ID,
    Email: user.Email,
    Name:  user.Name,
  }
}

// Use environment variables for secrets.
apiKey := os.Getenv("STRIPE_API_KEY")
if apiKey == "" {
  return errors.New("STRIPE_API_KEY not configured")
}
```

## Input Validation Patterns

### Schema Validation at Boundaries

```go
type CreateTaskInput struct {
  Title       string     `json:"title"`
  Description *string    `json:"description,omitempty"`
  Priority    string     `json:"priority,omitempty"`
  DueDate     *time.Time `json:"dueDate,omitempty"`
}

func (input CreateTaskInput) Validate() error {
  if strings.TrimSpace(input.Title) == "" {
    return errors.New("title is required")
  }
  if len(input.Title) > 200 {
    return errors.New("title must be 200 characters or fewer")
  }
  return nil
}

// Validate at the route handler.
func (handler *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
  var input CreateTaskInput
  if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
    writeJSON(w, http.StatusBadRequest, APIError{Error: ErrorBody{Code: "INVALID_JSON", Message: "Invalid input"}})
    return
  }
  if err := input.Validate(); err != nil {
    writeJSON(w, http.StatusUnprocessableEntity, APIError{Error: ErrorBody{Code: "VALIDATION_ERROR", Message: err.Error()}})
    return
  }
}
```

### File Upload Safety

```go
// Restrict file types and sizes.
var allowedTypes = map[string]struct{}{"image/jpeg": {}, "image/png": {}, "image/webp": {}}
const maxSize = 5 * 1024 * 1024 // 5MB

func validateUpload(file multipart.File, header *multipart.FileHeader) error {
	if _, ok := allowedTypes[header.Header.Get("Content-Type")]; !ok {
		return errors.New("file type not allowed")
	}
	if header.Size > maxSize {
		return errors.New("file too large (max 5MB)")
	}
	return nil
}
```

## Triaging govulncheck Results

Not all audit findings require immediate action. Use this decision tree:

```
govulncheck reports a vulnerability
├── Severity: critical or high
│   ├── Is the vulnerable code reachable in your app?
│   │   ├── YES --> Fix immediately (update, patch, or replace the dependency)
│   │   └── NO (dev-only dep, unused code path) --> Fix soon, but not a blocker
│   └── Is a fix available?
│       ├── YES --> Update to the patched version
│       └── NO --> Check for workarounds, consider replacing the dependency, or add to allowlist with a review date
├── Severity: moderate
│   ├── Reachable in production? --> Fix in the next release cycle
│   └── Dev-only? --> Fix when convenient, track in backlog
└── Severity: low
    └── Track and fix during regular dependency updates
```

**Key questions:**
- Is the vulnerable function actually called in your code path?
- Is the dependency a runtime dependency or dev-only?
- Is the vulnerability exploitable given your deployment context (e.g., a server-side vulnerability in a client-only app)?

When you defer a fix, document the reason and set a review date.

## Rate Limiting

```go
// General API rate limit.
apiLimiter := httprate.LimitByIP(100, 15*time.Minute)
router.With(apiLimiter).Route("/api", func(r chi.Router) {
	// ...
})

// Stricter limit for auth endpoints.
authLimiter := httprate.LimitByIP(10, 15*time.Minute)
router.With(authLimiter).Route("/api/auth", func(r chi.Router) {
	// ...
})
```

## Secrets Management

```
.env files:
  ├── .env.example  → Committed (template with placeholder values)
  ├── .env          → NOT committed (contains real secrets)
  └── .env.local    → NOT committed (local overrides)

.gitignore must include:
  .env
  .env.local
  .env.*.local
  *.pem
  *.key
```

**Always check before committing:**
```bash
# Check for accidentally staged secrets
git diff --cached | grep -i "password\|secret\|api_key\|token"
```

## Security Review Checklist

```markdown
### Authentication
- [ ] Passwords hashed with bcrypt/scrypt/argon2 (salt rounds ≥ 12)
- [ ] Session tokens are httpOnly, secure, sameSite
- [ ] Login has rate limiting
- [ ] Password reset tokens expire

### Authorization
- [ ] Every endpoint checks user permissions
- [ ] Users can only access their own resources
- [ ] Admin actions require admin role verification

### Input
- [ ] All user input validated at the boundary
- [ ] SQL queries are parameterized
- [ ] HTML output is encoded/escaped

### Data
- [ ] No secrets in code or version control
- [ ] Sensitive fields excluded from API responses
- [ ] PII encrypted at rest (if applicable)

### Infrastructure
- [ ] Security headers configured (CSP, HSTS, etc.)
- [ ] CORS restricted to known origins
- [ ] Dependencies audited for vulnerabilities
- [ ] Error messages don't expose internals
```
## See Also

For detailed security checklists and pre-commit verification steps, see `references/security-checklist.md`.

## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "This is an internal tool, security doesn't matter" | Internal tools get compromised. Attackers target the weakest link. |
| "We'll add security later" | Security retrofitting is 10x harder than building it in. Add it now. |
| "No one would try to exploit this" | Automated scanners will find it. Security by obscurity is not security. |
| "The framework handles security" | Frameworks provide tools, not guarantees. You still need to use them correctly. |
| "It's just a prototype" | Prototypes become production. Security habits from day one. |

## Red Flags

- User input passed directly to database queries, shell commands, or HTML rendering
- Secrets in source code or commit history
- API endpoints without authentication or authorization checks
- Missing CORS configuration or wildcard (`*`) origins
- No rate limiting on authentication endpoints
- Stack traces or internal errors exposed to users
- Dependencies with known critical vulnerabilities

## Verification

After implementing security-relevant code:

- [ ] `govulncheck ./...` shows no critical or high vulnerabilities
- [ ] No secrets in source code or git history
- [ ] All user input validated at system boundaries
- [ ] Authentication and authorization checked on every protected endpoint
- [ ] Security headers present in response (check with browser DevTools)
- [ ] Error responses don't expose internal details
- [ ] Rate limiting active on auth endpoints
