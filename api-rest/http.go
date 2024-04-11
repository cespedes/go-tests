package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

type hError struct {
	Status int
	Err    error
}

func (e hError) Error() string {
	return e.Err.Error()
}

func (e hError) Unwrap() error {
	return e.Err
}

func Error(e any) error {
	if err, ok := e.(hError); ok {
		return err
	}
	code := http.StatusBadRequest

	var err error
	if er, ok := e.(error); ok {
		err = er
	} else {
		err = errors.New(fmt.Sprint(e))
	}

	switch {
	case errors.Is(err, sql.ErrNoRows):
		code = http.StatusNotFound
	}
	return CodeError(code, e)
}

func CodeError(code int, e any) error {
	var err error
	if er, ok := e.(error); ok {
		err = er
	} else {
		err = errors.New(fmt.Sprint(e))
	}
	return hError{
		Status: code,
		Err:    err,
	}
}

// httpError sends a HTTP error as a response, with an optional HTTP status code.
// If code is not supplied, it defaults to BadRequest.
func httpError(w http.ResponseWriter, e any, codes ...int) {
	err := Error(e).(hError)

	if len(codes) > 0 {
		err.Status = codes[0]
	}

	httpMessage(w, err.Status, "error", err.Error())
}

// httpMessage sends a message as a JSON response
func httpMessage(w http.ResponseWriter, code int, label string, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{%q: %q}\n", label, msg)
}

// httpInfo sends an information message as a JSON response
func httpInfo(w http.ResponseWriter, msg any) {
	httpMessage(w, http.StatusOK, "info", fmt.Sprint(msg))
}
