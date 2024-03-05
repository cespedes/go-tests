package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: handle middleware
	log.Println("Server.ServeHTTP()")
	s.mux.ServeHTTP(w, r)
	// fmt.Fprintln(w, "Hello, world!")
}

func NewServer() *Server {
	var s Server
	s.mux = http.NewServeMux()
	return &s
}

type Request struct {
	Body []byte
}

func Get[Output any](s *Server, path string, f func(r *Request) (Output, error)) {
	log.Printf("Registering Get(%q, %v)\n", path, f)
	s.mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		var req Request

		log.Printf("Begin: GET %s\n", path)
		out, err := f(&req)
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}
		bb, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintf(w, "Encoding JSON: %v\n", err.Error())
			return
		}
		fmt.Fprintln(w, string(bb))
		log.Printf("End: GET %s: out=%v err=%v\n", path, out, err)
	})
}

func Post[Input, Output any](s *Server, path string, f func(r *Request, input Input) (Output, error)) {
	log.Printf("Registering Post(%q, %v)\n", path, f)
	s.mux.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		var req Request

		log.Printf("Begin: POST %s\n", path)
		var input Input
		out, err := f(&req, input)
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}
		bb, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintf(w, "Encoding JSON: %v\n", err.Error())
			return
		}
		fmt.Fprintln(w, string(bb))
		log.Printf("End: POST %s: out=%v err=%v\n", path, out, err)
	})
}
