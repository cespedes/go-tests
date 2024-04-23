package main

// Some ideas from:
// - https://github.com/go-chi/chi/blob/master/middleware/logger.go (github.com/go-chi/chi/v5/middleware/Logger()
// - github.com/MadAppGang/httplog

import (
	"fmt"
	"net/http"
	"time"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := NewWrapResponseWriter(w)
		t1 := time.Now()
		next.ServeHTTP(ww, r)
		t2 := time.Now()
		fmt.Printf("%s %d %s %s %s %s\n",
			t2.Format("2006-01-02 15:04:05"),
			ww.Status(),
			t2.Sub(t1),
			r.RemoteAddr, r.Method, r.URL.Path)
	})
}
