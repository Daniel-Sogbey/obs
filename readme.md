![CI](https://github.com/Daniel-Sogbey/obs/actions/workflows/ci.yml/badge.svg)

# 🔭 obs — Goroutine Observability for Go

`obs` is a lightweight structured observability toolkit for Go goroutines.

It helps you:

- Track goroutine lifecycle
- Model parent–child task relationships
- Detect slow or leaked tasks
- Visualize concurrency trees in real time
- Inspect state via an HTTP debug endpoint

Designed to be minimal, explicit, and production-friendly.

---

## Features

- Structured goroutine tracking via `context`
- Parent–child relationship modeling
- JSON debug endpoint
- CLI with live watch mode
- Tree visualization
- Slow task detection
- Leak detection (heuristic)
- Neon developer-friendly UI
- Zero heavy dependencies

---

## 📦 Installation

Install the library:

```bash
go get github.com/Daniel-Sogbey/obs
```

Build the CLI:

```bash
go build -o obs ./cmd/obs
```

---

# 🚀 Quick Start (HTTP Example)

Below is a minimal HTTP server example using `obs`.

## 1️⃣ Enable Observability

```go
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Daniel-Sogbey/obs/obs"
)

func main() {
	obs.Enable()
	obs.Listen(":7070")

	root := obs.With(context.Background(), "http-app")

	http.HandleFunc("/fast", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(root, "request /fast", 500*time.Millisecond)
		w.Write([]byte("fast response"))
	})

	http.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(root, "request /slow", 3*time.Second)
		w.Write([]byte("slow response"))
	})

	http.ListenAndServe(":8080", nil)
}

func handleRequest(parent context.Context, name string, delay time.Duration) {
	ctx := obs.With(parent, name)
	t := obs.FromContext(ctx)
	defer t.Done()

	t.MarkActive()
	time.Sleep(delay)
	t.MarkIdle()
}
```

---

## 2️⃣ Start Server

```bash
go run main.go
```

Visit:

```
http://localhost:8080/fast
http://localhost:8080/slow
```

---

## 3️⃣ Inspect Live Concurrency Tree

```bash
obs tree --watch
```

---

## Example Output

```
OBSERVABILITY
Tue, 17 Feb 2026 13:02:11 GMT

Active Goroutines: 3

http-app ● RUNNING 12.41s
  request /fast ● COMPLETED 501ms
  request /slow ● RUNNING 2.98s
```
---
![img](./docs/test_view.png)   
---

# Debug Endpoint

By default:

```
GET http://localhost:7070/debug/obs
```

Returns:

```json
[
  {
    "id": 1,
    "name": "http-app",
    "parent_id": 0,
    "state": "running",
    "duration": 12412500000
  }
]
```

---

# 🛠 CLI Commands

### Tree View

```bash
obs tree
```

### Live Tree

```bash
obs tree --watch
```

Optional refresh interval:

```bash
obs tree --watch --interval=1s
```

---

### Flat List

```bash
obs list
```

---

### Slow Goroutines

```bash
obs slow --threshold=2s
```

---

### Leak Detection (Heuristic)

```bash
obs leaks
```

---

# How It Works

- `obs.With()` creates a logical tracker
- Trackers are stored in a concurrent registry
- Parent–child relationships are derived from context propagation
- `Snapshot()` creates immutable state views
- CLI consumes snapshot JSON and builds a tree
- Watch mode re-renders the view periodically

No runtime hacks. No goroutine ID introspection.

This models structured concurrency explicitly.

---

# 📁 Project Structure

```
cmd/obs/       → CLI tool
obs/           → Library
examples/      → Demo usage
```

---

#  What This Is Not

- Not a Go scheduler inspector
- Not a `pprof` replacement
- Not runtime-level goroutine introspection

It tracks logical concurrent tasks you instrument.

---

#  Design Philosophy

- Explicit instrumentation
- Context-driven structure
- Minimal overhead
- Snapshot-based observability
- CLI-first experience
- Developer-friendly output

---

# 🤝 Contributing

PRs welcome.

If you add features, keep the core principles:

- No heavy dependencies
- No runtime magic
- Clean separation of concerns
- Production-safe behavior

---

Built with ❤️ for developers who want to *see* their concurrency.