package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Daniel-Sogbey/obs/obs"
)

func main() {
	// Enable observability
	obs.Enable()
	_ = obs.Listen(":7070") // debug endpoint

	// Root tracker
	rootCtx := obs.With(context.Background(), "http-app")

	mux := http.NewServeMux()

	mux.HandleFunc("/fast", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(rootCtx, w, r, 500*time.Millisecond)
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(rootCtx, w, r, 3*time.Second)
	})

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleRequest(parent context.Context, w http.ResponseWriter, r *http.Request, delay time.Duration) {
	// Each request gets its own tracker
	ctx := obs.With(parent, fmt.Sprintf("request %s", r.URL.Path))
	t := obs.FromContext(ctx)
	defer t.Done()

	t.MarkActive()

	time.Sleep(delay) // simulate processing

	t.MarkIdle()

	_, _ = w.Write([]byte("done\n"))
}
