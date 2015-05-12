// Package infuse provides an immutable, concurrency-safe middleware handler
// that conforms to http.Handler. An infuse.Handler is fully compatible with
// the Go standard library, supports flexible chaining, and provides a shared
// context between middleware handlers without relying on global state, locks,
// or shared closures.
package infuse

import "net/http"

// A Handler is a chained middleware handler that conforms to http.Handler.
type Handler interface {
	// Handle returns a copy of the current infuse.Handler with the provided
	// http.Handler attached. The provided http.Handler will be called when
	// the http.Handler attached before it calls infuse.Next. The first
	// http.Handler that is attached to an infuse.Handler will be called
	// when the infuse.Handler is served (via ServeHTTP).
	Handle(handler http.Handler) Handler

	// HandleFunc has the same behavior as Handle, but it takes a handler
	// function instead of an http.Handler.
	HandleFunc(handler func(http.ResponseWriter, *http.Request)) Handler

	// Stack has the same behavior as Handle but implicitly calls infuse.Next
	// after the provided http.Handler is called.
	//
	// Stack is useful for attaching generic handlers that do not know about
	// infuse and therefore do not call infuse.Next themselves. It can also be
	// used to attach one infuse.Handler to another infuse.Handler (as calls
	// to infuse.Next from an http.Handler attached to one infuse.Handler
	// do not "leak out" into a infuse.Handler nested above it).
	//
	// Do not call infuse.Next in a stacked http.Handler unless you intend to
	// serve the rest of the middleware chain more than once.
	Stack(handler http.Handler) Handler

	// StackFunc is the same as Stack, but it takes a handler function
	// instead of an http.Handler.
	StackFunc(handler func(http.ResponseWriter, *http.Request)) Handler

	// ServeHTTP serves the infuse.Handler, starting with the first
	// http.Handler attached.
	ServeHTTP(response http.ResponseWriter, request *http.Request)
}

// Next serves the next http.Handler in the middleware chain. It is always
// called from within an http.Handler that is handled by a infuse.Handler.
// The provided response must be same http.ResponseWriter provided to the
// current http.Handler in the chain.
//
// The boolean return value indicates whether the call succeeded. Next will
// return false if no subsequent http.Handler is available or if the response
// is invalid.
//
// Calling Next multiple times in the same handler will call all remaining
// http.Handlers in the middleware chain each time.
func Next(response http.ResponseWriter, request *http.Request) bool {
	sharedResponse, ok := response.(infuseResponse)
	return ok && sharedResponse.next(request)
}

// New returns a new infuse.Handler.
func New() Handler {
	return (*layer)(nil)
}

type layer struct {
	handler http.Handler
	prev    *layer
}

func (l *layer) Handle(handler http.Handler) Handler {
	return &layer{handler, l}
}

func (l *layer) HandleFunc(handler func(http.ResponseWriter, *http.Request)) Handler {
	return l.Handle(http.HandlerFunc(handler))
}

func (l *layer) Stack(handler http.Handler) Handler {
	return l.HandleFunc(func(response http.ResponseWriter, request *http.Request) {
		handler.ServeHTTP(response, request)
		Next(response, request)
	})
}

func (l *layer) StackFunc(handler func(http.ResponseWriter, *http.Request)) Handler {
	return l.Stack(http.HandlerFunc(handler))
}

func (l *layer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if l == nil {
		return
	}

	sharedResponse := newLayeredResponse(response)
	current := l
	for ; current.prev != nil; current = current.prev {
		sharedResponse.layers = append(sharedResponse.layers, current)
	}
	current.handler.ServeHTTP(sharedResponse.extend(), request)
}
