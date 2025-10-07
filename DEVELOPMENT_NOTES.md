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
✅ **Status code range checking (200-299)** - COMPLETED!
✅ **Proper URL validation** - COMPLETED!
✅ **Context-based cancellation** - COMPLETED!
✅ **Configured interval usage** - COMPLETED!
✅ **JSON field case mismatch** - COMPLETED!
✅ **Basic metrics implementation** - IN PROGRESS (debugging needed)

## Critical Issues Fixed/To Fix

### ✅ Fixed: WaitGroup Bug
```go
// BEFORE (BROKEN):
wg.Add(i)  // This adds 0, 1, 2, 3... 

// AFTER (CORRECT):
wg.Add(1)  // Add 1 for each goroutine
```

### ✅ COMPLETED: All 5 High Priority Issues!
1. ✅ **Status Code Handling**: Now properly checks 200-299 range for healthy
2. ✅ **URL Validation**: Proper `net/url` parsing with scheme and host validation
3. ✅ **Signal Handling**: Clean context-based cancellation pattern
4. ✅ **Config Interval**: Using `h.config.IntervalSeconds` throughout
5. ✅ **JSON Case Mismatch**: Fixed `json:"urls"` tag to match JSON file

### 🔧 Current Issue: Metrics Debugging
**Problem**: Metrics showing incorrect values:
- Average response times in millions of milliseconds (should be ~50ms)
- Multiple metric entries per URL (possible key inconsistency)
- Some URLs showing 0 successful checks despite 200 responses

**Suspected Issues**:
- Time units mixing (`time.Milliseconds()` vs `time.Duration` math)
- Map key inconsistencies (URL formatting differences)
- Possible concurrent access issues with map[string]URLMetrics

**Next Debug Steps**:
- Add URL to metrics log output to identify which URL each line represents
- Check for trailing slashes or other URL key variations
- Verify time calculation units in average computation

### Medium Priority
- Add max URLs limit (safety: max 100 URLs)
- Improve error handling with more context
- Better graceful shutdown

### Low Priority
- Unit tests
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

## Quick Reference: Status Code Logic ✅ IMPLEMENTED
```go
// BEFORE (incomplete)
switch resp.StatusCode {
case http.StatusOK:
    // healthy
case http.StatusInternalServerError:
    // unhealthy
}

// AFTER (correct range checking)
if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    // healthy - covers all 2xx codes
} else {
    // unhealthy
}
```

## Quick Reference: URL Validation ✅ IMPLEMENTED
```go
// BEFORE (stub)
func validateUrl(url string) bool {
    return true
}

// AFTER (proper validation)
func validateUrl(u string) bool {
    parsedUrl, err := url.Parse(u)
    if err != nil {
        slog.Error("Unable to parse URL", "err", err)
        return false
    }
    if parsedUrl.Host == "" { return false }
    if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" { return false }
    return true
}
```

## Metrics Implementation Lessons Learned

### Concurrent Map Access Patterns
```go
// PROBLEM: Cannot take address of map values
m := &h.metrics[url]  // ERROR: cannot take address

// SOLUTION 1: Store pointers in map
metrics map[string]*URLMetrics

// SOLUTION 2: Access directly
h.metrics[url].Mutex.Lock()
m := h.metrics[url]
// ... modify m ...
h.metrics[url] = m  // Put back
```

### RWMutex Patterns
```go
// READ operations (multiple readers OK)
m.Mutex.RLock()
value := m.TotalChecks
m.Mutex.RUnlock()

// WRITE operations (exclusive access)
m.Mutex.Lock()
m.TotalChecks += 1
m.Mutex.Unlock()
```

### Common Gotchas
- **Never defer unlock in loops** - accumulates locks!
- **Copy vs pointer semantics** - copying a struct copies the mutex too
- **Time arithmetic** - mixing `time.Duration` and `int64` milliseconds

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
✅ **Context patterns and cancellation** - Clean signal-to-context flow
✅ **URL parsing and validation** - Using `net/url` package effectively
✅ **HTTP status code semantics** - Understanding 2xx range for success
✅ **JSON struct tags** - Proper mapping between JSON and Go structs
✅ **Concurrent programming basics** - Goroutines, WaitGroups, and coordination

## Learning Goals In Progress
🔄 **Concurrent state management** - Metrics with mutex/RWMutex patterns
🔄 **Map vs pointer semantics** - Learning when you can/cannot take addresses
🔄 **Time/Duration arithmetic** - Proper handling of time calculations
🔄 Proper error handling strategies
🔄 HTTP client best practices

## Current Session Progress (MAJOR WIN! 🎉)
**COMPLETED ALL 5 HIGH PRIORITY ISSUES:**
1. ✅ Status code range checking (200-299)
2. ✅ URL validation with proper scheme/host checking
3. ✅ Context-based signal handling (removed unnecessary sigChan parameter)
4. ✅ Configured interval usage (no more hardcoded 1 second)
5. ✅ JSON case mismatch fix

**STARTED: Basic Metrics Implementation**
- Learned about concurrent map access issues
- Explored mutex vs RWMutex for different access patterns
- Discovered Go's restriction on taking addresses of map values
- Implemented per-URL metrics tracking structure

## Next Session Priorities
1. **Debug metrics calculation issues** (time units, URL keys, concurrent access)
2. **Finalize metrics display** (clean up output formatting)
3. **Add max URLs limit** (safety: max 100 URLs)
4. **Enhanced error handling** with more context
5. **Unit tests** for validation functions

## Development Philosophy Notes
- "Tiger Style" programming - account for negative space, program in assumptions
- Learning without AI assistance to build stronger fundamentals
- Start simple, add complexity only when needed
- Follow Go idioms and conventions
