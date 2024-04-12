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

func Error(f any, a ...any) error {
	if err, ok := f.(hError); ok {
		return err
	}

	var err error
	if e, ok := f.(error); ok {
		err = e
	} else if s, ok := f.(string); ok {
		err = fmt.Errorf(s, a...)
	} else {
		err = errors.New(fmt.Sprint(f))
	}

	code := http.StatusBadRequest
	switch {
	case errors.Is(err, sql.ErrNoRows):
		code = http.StatusNotFound
	}
	return CodeError(code, err)
}

func CodeError(code int, f any, a ...any) error {
	var err error
	if e, ok := f.(error); ok {
		err = e
	} else if s, ok := f.(string); ok {
		err = fmt.Errorf(s, a...)
	} else {
		err = errors.New(fmt.Sprint(f))
	}

	return hError{
		Status: code,
		Err:    err,
	}
}

// httpError sends a HTTP error as a response.
// The default HTTP status code is BadRequest.
func httpError(w http.ResponseWriter, f any, a ...any) {
	err := Error(f, a...).(hError)

	httpMessage(w, err.Status, "error", err.Error())
}

// httpError sends a HTTP error as a response, with a specified HTTP status code.
func httpCodeError(w http.ResponseWriter, code int, f any, a ...any) {
	err := CodeError(code, f, a...).(hError)

	httpError(w, err)
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
