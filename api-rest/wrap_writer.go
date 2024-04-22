package main

// Derived from: https://github.com/go-chi/chi/tree/master/middleware

import (
	"io"
	"net/http"
)

// NewWrapResponseWriter wraps an http.ResponseWriter, returning a proxy that allows you to
// see the returned status code, bytes written, and content written.
func NewWrapResponseWriter(w http.ResponseWriter) *WrapResponseWriter {
	wr := WrapResponseWriter{ResponseWriter: w}

	return &wr
}

// WrapResponseWriter is a proxy around an http.ResponseWriter that allows you to
// see the returned status code, bytes written, and content written.
type WrapResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
	bytes       int
	tee         io.Writer
}

// WriteHeader calls the ResponseWriter's WriteHeader
// after writing doen the code.
func (b *WrapResponseWriter) WriteHeader(code int) {
	if !b.wroteHeader {
		b.code = code
		b.wroteHeader = true
		b.ResponseWriter.WriteHeader(code)
	}
}

// WriteHeader calls the ResponseWriter's Write,
// updates the written bytes and optionally calls the io.Writer
// used in a previous Tee.
func (b *WrapResponseWriter) Write(buf []byte) (int, error) {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
	n, err := b.ResponseWriter.Write(buf)
	if b.tee != nil {
		_, err2 := b.tee.Write(buf[:n])
		// Prefer errors generated by the proxied writer.
		if err == nil {
			err = err2
		}
	}
	b.bytes += n
	return n, err
}

// Status returns the HTTP status of the request, or 0 if one has not
// yet been sent.
func (b *WrapResponseWriter) Status() int {
	return b.code
}

// BytesWritten returns the total number of bytes sent to the client.
func (b *WrapResponseWriter) BytesWritten() int {
	return b.bytes
}

// Tee causes the response body to be written to the given io.Writer in
// addition to proxying the writes through. Only one io.Writer can be
// tee'd to at once: setting a second one will overwrite the first.
// Writes will be sent to the proxy before being written to this
// io.Writer. It is illegal for the tee'd writer to be modified
// concurrently with writes.
func (b *WrapResponseWriter) Tee(w io.Writer) {
	b.tee = w
}
