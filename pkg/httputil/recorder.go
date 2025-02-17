package httputil

import (
	"bytes"
	"net/http"
)

// ResponseRecorderWriter is an implementation of http.ResponseWriter that
// can also records response body
type ResponseRecorderWriter struct {
	writer http.ResponseWriter
	Body   *bytes.Buffer
}

// NewRespRecorderWriter returns an initialized ResponseRecorderWriter
func NewRespRecorderWriter(w http.ResponseWriter) *ResponseRecorderWriter {
	return &ResponseRecorderWriter{
		writer: w,
		Body:   new(bytes.Buffer),
	}
}

// Header implements http.ResponseWriter. It returns the response
// headers to mutate within a handler.
func (w *ResponseRecorderWriter) Header() http.Header {
	return w.writer.Header()
}

// Write implements http.ResponseWriter. The data in buf is copied to
// w.Body and then pass to the real ResponseWriter.
func (w *ResponseRecorderWriter) Write(bytes []byte) (int, error) {
	w.Body.Write(bytes[0:])
	return w.writer.Write(bytes)
}

// WriteHeader implements http.ResponseWriter.
func (w *ResponseRecorderWriter) WriteHeader(i int) {
	w.writer.WriteHeader(i)
}
