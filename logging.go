package main

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps an http.ResponseWriter to capture the status code
// written by downstream handlers, since http.ResponseWriter has no getter.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// accessLogMiddleware logs one line per HTTP request: remote address,
// method, path, response status, and how long it took to handle.
func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		log.Printf("access: remote=%s method=%s path=%s status=%d duration=%s",
			r.RemoteAddr, r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
