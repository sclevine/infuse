// Package infuse provides an immutable, concurrency-safe middleware handler
// that conforms to http.Handler. An infuse.Handler is fully compatible with
// the standard library, supports flexible chaining, and provides a shared
// context between middleware handlers within a single request without
// relying on global state, locks, or closures.
package infuse

import "net/http"

// A Handler is an immutable, thread-safe middleware handler that conforms to
// http.Handler. Handlers are fully compatible with the standard library,
// support flexible chaining, and provide shared contexts for individual
// requests without relying on global state, locks, or closures.
//
// An http.Handler passed to the Handle methods can call the http.Handler
// provided immediately after it using infuse.Next. All handlers within a
// single request-response cycle can share data of any type using infuse.Get
// and infuse.Set.
type Handler interface {
	// Handle returns a new infuse.Handler that serves the provided
	// http.Handler after all existing http.Handlers are served. The
	// newly-provided http.Handler will be called when the previously-provided
	// http.Handler calls infuse.Next. The provided http.Handler may be
	// an infuse.Handler.
	Handle(handler http.Handler) Handler

	// HandleFunc is the same as Handle, but it takes a handler function
	// instead of an http.Handler.
	HandleFunc(handler func(http.ResponseWriter, *http.Request)) Handler

	// ServeHTTP serves the first http.Handler provided to the infuse.Handler.
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

func (l *layer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if l == nil {
		return
	}

	sharedResponse := convertResponse(response)

	current := l
	for ; current.prev != nil; current = current.prev {
		sharedResponse.layers = append(sharedResponse.layers, current)
	}
	current.handler.ServeHTTP(sharedResponse, request)
}

func convertResponse(response http.ResponseWriter) *contextualResponse {
	converted, ok := response.(*contextualResponse)
	if !ok {
		return &contextualResponse{ResponseWriter: response}
	}
	return converted
}

type contextualResponse struct {
	http.ResponseWriter
	context interface{}
	layers  []*layer
}

// Next should only be called from within an http.Handler that is being served
// by an infuse.Handler. It will serve the next http.Handler in the middleware
// stack. Next returns false if no subsequent http.Handler is available, or if
// the provided response object was not served by an infuse.Handler. Calling
// Next multiple times in the same handler will run the rest of the middleware
// stack for each call.
//
// Next must be called with the same response object provided to the
// caller. Like http.ResponseWriter, Next is not safe for concurrent usage
// with other parts of the same request-response cycle.
func Next(response http.ResponseWriter, request *http.Request) bool {
	sharedResponse := convertResponse(response)
	layers := sharedResponse.layers
	if len(layers) == 0 {
		return false
	}

	next := layers[len(layers)-1]
	sharedResponse.layers = layers[:len(layers)-1]
	defer func() { sharedResponse.layers = layers }()

	next.handler.ServeHTTP(response, request)
	return true
}

// Get will retrieve a value that is shared by the every infuse-served
// http.Handler in the same request-response cycle. Any changes to data
// pointed to by the returned value will be seen by other http.Handlers
// that call infuse.Get in the same request-response cycle.
//
// To retrieve a value of a particular type, wrap infuse.Get as such:
//   func GetMyContext(response http.ResponseWriter) *MyContext {
//      return infuse.Get(response).(*MyContext)
//   }
//
// Example using a map:
//   func GetMyMap(response http.ResponseWriter) map[string]string {
//      return infuse.Get(response).(map[string]string)
//   }
//
//   func CreateMyMap(response http.ResponseWriter) {
//      infuse.Set(response, make(make[string]string))
//   }
func Get(response http.ResponseWriter) interface{} {
	return response.(*contextualResponse).context
}

// Set will store a value that will be shared by the every infuse-served
// http.Handler in the same request-response cycle.
func Set(response http.ResponseWriter, value interface{}) {
	response.(*contextualResponse).context = value
}
