package infuse

import (
	"io"
	"net/http"
)

type infuseResponse interface {
	next(request *http.Request) bool
	get() interface{}
	set(value interface{})
}

type httpResponse interface {
	http.CloseNotifier
	http.Flusher
	http.Hijacker
	io.ReaderFrom
	stringWriter
}

type layeredResponse struct {
	*contextualResponse
	layers []*layer
}

func newLayeredResponse(response http.ResponseWriter) *layeredResponse {
	return &layeredResponse{&contextualResponse{response, nil}, nil}
}

func (l *layeredResponse) next(request *http.Request) bool {
	if len(l.layers) == 0 {
		return false
	}

	next := l.layers[len(l.layers)-1]
	remaining := l.layers[:len(l.layers)-1]
	sharedResponse := &layeredResponse{l.contextualResponse, remaining}
	next.handler.ServeHTTP(sharedResponse.extend(), request)
	return true
}

// extend detects if the underlying response is a *http.response or
// *httptest.ResponseRecorder and returns the *layeredResponse extended with
// any extra methods defined on those types. This allows a http.ResponseWriter
// provided to handlers to be type-asserted into an http.Flusher,
// http.CloseNotifier, http.Hijacker, etc.
func (l *layeredResponse) extend() http.ResponseWriter {
	if _, ok := l.ResponseWriter.(httpResponse); ok {
		return &fullResponse{l}
	}
	if _, ok := l.ResponseWriter.(http.Flusher); ok {
		return &flushableResponse{l}
	}
	return l
}
