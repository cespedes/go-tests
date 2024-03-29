package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Status int
	Err    error
}

func (e Error) Error() string {
	return e.Err.Error()
}

func (e Error) Unwrap() error {
	return e.Err
}

func NewError(code int, err error) error {
	return Error{
		Status: code,
		Err:    err,
	}
}

// httpError sends a HTTP error as a response, with an optional HTTP status code.
// If code is not supplied, it defaults to InternalServerError.
func httpError(w http.ResponseWriter, err any, codes ...int) {
	code := http.StatusInternalServerError

	if err == sql.ErrNoRows {
		code = http.StatusNotFound
	}

	if er, ok := err.(error); ok {
		var e Error
		if errors.As(er, &e) {
			code = e.Status
		}
	}

	if len(codes) > 0 {
		code = codes[0]
	}

	httpMessage(w, code, "error", fmt.Sprint(err))
}

// httpMessage sends
func httpMessage(w http.ResponseWriter, code int, label string, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{%q: %q}\n", label, msg)
}
