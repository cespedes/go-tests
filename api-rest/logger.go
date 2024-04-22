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
	fmt.Println("outside: logger()")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		fmt.Println("logger: before next()")
		next.ServeHTTP(w, r)
		fmt.Println("logger: after next()")
		t2 := time.Now()
		fmt.Printf("%s %s %s %s %s\n",
			t2.Format("2006-01-02 15:04:05"),
			t2.Sub(t1),
			r.RemoteAddr, r.Method, r.URL.Path)
	})
}
