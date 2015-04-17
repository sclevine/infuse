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
	current.handler.ServeHTTP(sharedResponse, request)
}

type layeredResponse struct {
	*contextualResponse
	layers []*layer
}

type contextualResponse struct {
	http.ResponseWriter
	context interface{}
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
	next.handler.ServeHTTP(sharedResponse, request)
	return true
}

// Next serves the next http.Handler in the middleware chain. Next uses the
// provided http.ResponseWriter to determine which http.Handler to call, so
// the response must be same http.ResponseWriter provided to the current
// http.Handler in the chain.
//
// The boolean return value indicates whether the call succeeded. Next will
// return false if no subsequent http.Handler is available.
//
// Calling Next multiple times in the same handler will call all remaining
// http.Handlers in the middleware chain for each call.
func Next(response http.ResponseWriter, request *http.Request) bool {
	sharedResponse, ok := response.(*layeredResponse)
	return ok && sharedResponse.next(request)
}

// Get will retrieve a context value that is shared by all http.Handlers
// attached to the same infuse.Handler. The context value is associated with
// an http.ResponseWriter, so it has the same life cycle as the provided
// response. Get will return nil if the provided response does not have an
// associated context value.
//
// If the context value is a pointer, map, or slice, then changes to the data
// in one http.Handler will be seen by other http.Handlers that share it.
//
// To retrieve a value of a particular type, consider wrapping Get and Set in
// another function. For example, this function could be used to share a map
// of floating-point numbers between middleware handlers:
//
//   func HandlerMap(response http.ResponseWriter) map[string]float64 {
//      handlerMap, ok := infuse.Get(response).(map[string]float64)
//      if !ok || handlerMap == nil {
//         handlerMap = make(map[string]float64)
//         if ok := infuse.Set(response, handlerMap); !ok {
//            return nil
//         }
//      }
//      return handlerMap
//   }
func Get(response http.ResponseWriter) interface{} {
	sharedResponse, ok := response.(*layeredResponse)
	if !ok {
		return nil
	}
	return sharedResponse.context
}

// Set will store a context value that is shared by all http.Handlers attached
// to the same infuse.Handler. The context value is associated with an
// http.ResponseWriter, so it has the same life cycle as the provided response.
//
// The boolean return value indicates whether the setting the context value
// succeeded. Set will return false if the provided response is invalid.
func Set(response http.ResponseWriter, value interface{}) bool {
	sharedResponse, ok := response.(*layeredResponse)
	if !ok {
		return false
	}
	sharedResponse.context = value
	return true
}
