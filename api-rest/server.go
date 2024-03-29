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
	Mux *http.ServeMux
}

// Request wraps http.Request, offering all its data, and some more.
type Request struct {
	*http.Request
}

// ServeHTTP dispatches the request to the handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: handle middleware
	log.Printf("Server.ServeHTTP(%s %s)", r.Method, r.URL)
	s.Mux.ServeHTTP(w, r)
	// fmt.Fprintln(w, "Hello, world!")
}

// NewServer allocates and returns a new Server.
func NewServer() *Server {
	var s Server
	s.Mux = http.NewServeMux()
	return &s
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

// HandleOutFunc returns a HTTP handler that calls a function and encodes its output as a JSON response.
func HandleOutFunc[Output any](
	s *Server,
	f func(ctx context.Context, s *Server, r *Request) (Output, error),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := Request{
			Request: r,
		}

		out, err := f(context.Background(), s, &req)
		if err != nil {
			httpError(w, err)
			return
		}
		outJSON(w, out)
	})
}

// HandleInOutFunc returns a HTTP handler that decodes a JSON input,
// calls a function and encodes its output as a JSON response.
func HandleInOutFunc[Input, Output any](
	s *Server,
	f func(ctx context.Context, s *Server, r *Request, input Input) (Output, error),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := Request{
			Request: r,
		}

		var input Input

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			httpError(w, fmt.Errorf("parsing input: %w", err), http.StatusBadRequest)
			return
		}

		out, err := f(context.Background(), s, &req, input)
		if err != nil {
			httpError(w, err)
			return
		}
		outJSON(w, out)
	})
}
