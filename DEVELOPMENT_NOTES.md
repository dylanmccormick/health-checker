# Health Checker Development Notes

## Project Overview
16 Go projects in 16 weeks - Project #2: URL Health Checker
Learning structured logging, HTTP clients, concurrency, and "Tiger Style" programming without AI assistance.

## Current Status
✅ Basic structure implemented with structured logging (slog + tint)
✅ Configuration management with JSON
✅ Signal handling setup
✅ Goroutine foundation
✅ HTTP client setup

## Critical Issues Fixed/To Fix

### ✅ Fixed: WaitGroup Bug
```go
// BEFORE (BROKEN):
wg.Add(i)  // This adds 0, 1, 2, 3... 

// AFTER (CORRECT):
wg.Add(1)  // Add 1 for each goroutine
```

### 🔧 High Priority Remaining Issues
1. **Status Code Handling**: Only checking 200/500, need 200-299 = healthy
2. **URL Validation**: `validateUrl()` just returns `true` - needs actual validation
3. **Signal Handling**: Context-based cancellation instead of direct signal channels
4. **Config Interval**: Using hardcoded 1 second instead of `config.IntervalSeconds`
5. **JSON Case Mismatch**: `"urls"` in JSON vs `Urls` in struct

### Medium Priority
- Add max URLs limit (safety: max 100 URLs)
- Improve error handling with more context
- Better graceful shutdown

### Low Priority
- Unit tests
- Metrics (success rates, avg response times)
- Enhanced logging

## Context Best Practices Learned

### ✅ DO: Pass Context as First Parameter
```go
func HealthCheck(ctx context.Context, url string, c Config) error {
    // ...
}
```

### ❌ DON'T: Store Context in Structs
```go
// AVOID
type Checker struct {
    ctx context.Context  // Don't do this
}
```

### Context Flow Pattern
```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handle signals
    go func() {
        <-sigChan
        cancel() // Broadcasts to all goroutines
    }()
    
    run(ctx, config)
}

func run(ctx context.Context, c Config) {
    for _, url := range c.Urls {
        go healthCheck(ctx, url, c) // Pass down the stack
    }
}

func healthCheck(ctx context.Context, url string, c Config) {
    for {
        select {
        case <-ctx.Done(): // Always check for cancellation
            return
        case <-ticker.C:
            // Do work
        }
    }
}
```

## HTTP Client + Context Patterns

### Two Types of Timeouts
1. **Client-Level Timeout**: `http.Client{Timeout: 5*time.Second}`
2. **Request-Level Context**: `http.NewRequestWithContext(ctx, "GET", url, nil)`

**The shorter timeout wins!**

### Recommended Pattern
```go
// Always use context with HTTP requests
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    return err
}

resp, err := client.Do(req)
if err != nil {
    // Check if context was cancelled vs network error
    if ctx.Err() != nil {
        // Context cancelled - expected during shutdown
        return nil
    }
    // Actual network error
    return err
}
```

## Architecture Decision: Functions vs Structs

### ✅ Current Choice: Functions (Correct for this project)
**Why functions are right here:**
- Single, focused purpose
- Simple configuration that doesn't change
- Stateless operations
- Small scope

### When to Use Structs Instead
Only when you need:
- State management (tracking stats between calls)
- Multiple related operations (Start/Stop/GetStats/AddURL)
- Configuration that changes at runtime
- Complex dependencies (database, alerting, metrics)

**Principle**: Start simple, refactor when complexity demands it.

## Current Code Structure
```
health-checker/
├── main.go          # Entry point, signal handling, logging setup
├── config.go        # Config loading and validation
├── checker.go       # Health check logic (mostly empty)
├── config.json      # URLs and settings
├── assert/
│   └── assert.go    # Custom assertion helper
└── go.mod           # Dependencies (tint for colored logging)
```

## Key Dependencies
- `github.com/lmittmann/tint` - Colored structured logging
- Standard library: `log/slog`, `context`, `net/http`, `os/signal`

## Quick Reference: Status Code Logic
```go
// Current (incomplete)
switch resp.StatusCode {
case http.StatusOK:
    // healthy
case http.StatusInternalServerError:
    // unhealthy
}

// Should be (per requirements)
healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
```

## Quick Reference: URL Validation
```go
// Current (stub)
func validateUrl(url string) bool {
    return true
}

// Should be
func validateUrl(urlStr string) bool {
    u, err := url.Parse(urlStr)
    return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}
```

## Testing Strategy
- Unit tests for config validation
- Unit tests for URL validation  
- Integration tests for HTTP health checks
- Test graceful shutdown behavior

## Learning Goals Achieved
✅ Structured logging with slog
✅ Configuration management
✅ Basic goroutine patterns
✅ Signal handling concepts
✅ HTTP client setup

## Learning Goals In Progress
🔄 Context patterns and cancellation
🔄 Proper error handling strategies
🔄 Goroutine coordination with WaitGroups
🔄 HTTP client best practices

## Next Session Priorities
1. Fix status code range checking (200-299)
2. Implement proper URL validation
3. Convert signal handling to context-based cancellation
4. Use configured interval instead of hardcoded 1 second
5. Fix JSON field case mismatch

## Development Philosophy Notes
- "Tiger Style" programming - account for negative space, program in assumptions
- Learning without AI assistance to build stronger fundamentals
- Start simple, add complexity only when needed
- Follow Go idioms and conventions