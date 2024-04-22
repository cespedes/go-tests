package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Server represents a server.
type Server struct {
	once        sync.Once
	root        http.Handler
	Mux         *http.ServeMux
	Middlewares []func(http.Handler) http.Handler
}

// Request wraps http.Request, offering all its data, and some more.
type Request struct {
	*Server
	*http.Request
}

// ServeHTTP dispatches the request to the handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.once.Do(func() {
		log.Println("Server.ServeHTTP: first time: chaining middlewares")
		s.root = s.Mux
		for i := len(s.Middlewares) - 1; i >= 0; i-- {
			s.root = s.Middlewares[i](s.root)
		}
	})
	log.Printf("Server.ServeHTTP(%s %s)", r.Method, r.URL)
	s.root.ServeHTTP(w, r)
	// fmt.Fprintln(w, "Hello, world!")
}

// AddMiddleware adds a new middleware to the server
func (s *Server) AddMiddleware(f func(next http.Handler) http.Handler) {
	s.Middlewares = append(s.Middlewares, f)
}

type contextServerKey struct{}

// NewServer allocates and returns a new Server.
func NewServer() *Server {
	var s Server
	s.Mux = http.NewServeMux()

	// This adds a middleware that adds the server struct
	// to the context of all the requests.
	s.Middlewares = append(s.Middlewares, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextServerKey{}, &s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	s.Middlewares = append(s.Middlewares, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("called middlewares[0]")
			next.ServeHTTP(w, r)
		})
	})
	s.Middlewares = append(s.Middlewares, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("called middlewares[1]")
			next.ServeHTTP(w, r)
		})
	})
	return &s
}

// NewRequest creates a new Request from a http.Request, ready to use.
func NewRequest(r *http.Request) *Request {
	return &Request{
		Request: r,
		Server:  r.Context().Value(contextServerKey{}).(*Server),
	}
}

// outJSON writes the JSON-encoded object v to the http.ResponseWriter w.
func outJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(w)
	err := e.Encode(v)
	if err != nil {
		httpError(w, err)
		log.Printf("Encoding JSON response: %v", err)
	}
}

// A None type will not be decoded as a body input.
type None struct{}

// Handle returns a HTTP handler that decodes a JSON input,
// calls a function and encodes its output as a JSON response.
func Handle[Input, Output any](
	f func(*Request, Input) (Output, error),
	permFuncs ...func(*Request) bool,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := NewRequest(r)
		for _, p := range permFuncs {
			if !p(req) {
				httpError(w, Error("permission denied"))
				return
			}
		}

		var input Input

		if _, ok := any(input).(None); !ok {
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&input); err != nil {
				httpError(w, "parsing input: %w", err)
				return
			}
		}

		out, err := f(req, input)

		if err != nil {
			httpError(w, err)
			return
		}

		var o any = out

		// if the returned type is a string, output it as a "info" message:
		if s, ok := o.(string); ok {
			httpInfo(w, s)
			return
		}

		// if the returned type is a []byte, output it directly:
		if b, ok := o.([]byte); ok {
			w.Write(b)
			return
		}

		outJSON(w, out)
	})
}

// permission functions
func OnlyRoot(r *Request) bool {
	return r.Header.Get("auth") == "root"
}
