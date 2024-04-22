package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Server represents a server.
type Server struct {
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
	log.Printf("Server.ServeHTTP(%s %s)", r.Method, r.URL)
	var m http.Handler
	m = s.Mux
	for i := len(s.Middlewares) - 1; i >= 0; i-- {
		m = s.Middlewares[i](m)
	}
	m.ServeHTTP(w, r)
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

func handleBefore(r *http.Request, permFuncs ...func(*Request) bool) (*Request, error) {
	req := NewRequest(r)
	for _, p := range permFuncs {
		if !p(req) {
			return nil, fmt.Errorf("permission denied")
		}
	}
	return req, nil
}

func handleAfter(w http.ResponseWriter, out any, err error) {
	if err != nil {
		httpError(w, err)
		return
	}

	// if the returned type is a string, output it as a "info" message:
	if s, ok := out.(string); ok {
		httpInfo(w, s)
		return
	}

	outJSON(w, out)
}

// HandleOut returns a HTTP handler that calls a function and encodes its output as a JSON response.
func HandleOut[Output any](
	f func(*Request) (Output, error),
	permFuncs ...func(*Request) bool,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := handleBefore(r, permFuncs...)
		if err != nil {
			httpError(w, err)
			return
		}

		out, err := f(req)

		handleAfter(w, out, err)
	})
}

// HandleInOut returns a HTTP handler that decodes a JSON input,
// calls a function and encodes its output as a JSON response.
func HandleInOut[Input, Output any](
	f func(*Request, Input) (Output, error),
	permFuncs ...func(*Request) bool,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := handleBefore(r, permFuncs...)
		if err != nil {
			httpError(w, err)
			return
		}

		var input Input

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			httpError(w, "parsing input: %w", err)
			return
		}

		out, err := f(req, input)

		handleAfter(w, out, err)
	})
}

// Handle returns a HTTP handler that decodes a JSON input,
// calls a function and encodes its output as a JSON response.
func Handle[Input any](
	f func(*Request, Input) (any, error),
	permFuncs ...func(*Request) bool,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := handleBefore(r, permFuncs...)
		if err != nil {
			httpError(w, err)
			return
		}

		var input Input

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			httpError(w, "parsing input: %w", err)
			return
		}

		out, err := f(req, input)

		handleAfter(w, out, err)
	})
}

// permission functions
func OnlyRoot(r *Request) bool {
	return r.Header.Get("auth") == "root"
}
