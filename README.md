# 16 Projects in 16 weeks. Project #2: URL Health Checker
This is a project to learn a bit more about structured logging and calling endpoints with golang. Trying to assert some more Tiger Style programming into my project where I account for negative space and program in my assumptions.

I'm trying to learn what I can without using AI to assist so when the AI slop dominates the marketplace, I can fix it :)

## Requirements of this project:

### Must Have
- Read a list of URLs from a config file (JSON or YAML - your choice)
- Perform HTTP GET requests to each URL concurrently using goroutines
- Check response status codes (200-299 = healthy, anything else = unhealthy)
- Measure response time for each request
- Display results to stdout with structured logging
- Run checks periodically (e.g., every 30 seconds) using time.Ticker
- Graceful shutdown on SIGINT (Ctrl+C)

### Technical Requirements
Concurrency:

- Use goroutines to check multiple URLs simultaneously
- Use channels to collect results from goroutines
- Set a reasonable timeout on HTTP requests (e.g., 5 seconds)

Error Handling:

- Handle network errors explicitly
- Handle timeout errors
- Log errors with context (which URL failed, why)

Configuration:

- URLs should be in a config file (not hardcoded)
- Make check interval configurable
- Make HTTP timeout configurable

Safety:

- Validate all URLs before starting checks
- Set upper bounds on concurrent checks (e.g., max 100 URLs)
- Explicit resource cleanup on shutdown

Performance:

- Concurrent checks (don't block on one slow URL)
- Reuse HTTP client (don't create new one per request)

Developer Experience:

- Clear, descriptive variable names with units: timeoutSeconds, checkIntervalSeconds
- Structured logging with log/slog
- Document why you chose specific timeout values
