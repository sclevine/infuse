// Package infuse provides an immutable, concurrency-safe middleware handler
// that conforms to http.Handler. An infuse.Handler is fully compatible with
// the standard library, supports flexible chaining, and provides a shared
// context between middleware handlers within a single request without
// relying on global state, locks, or shared closures.
package infuse

import "net/http"

// A Handler is a middleware handler that conforms to http.Handler.
//
// An http.Handler passed to the Handle method can call the http.Handler
// provided immediately after it using infuse.Next. All handlers within a
// single request-response cycle can share data of any type using infuse.Get
// and infuse.Set.
type Handler interface {
	// Handle returns a new infuse.Handler that serves the provided
	// http.Handler after all existing http.Handlers attached to the old
	// infuse.Handler are served. The provided http.Handler will be called
	// if and only if the previously-provided http.Handler calls infuse.Next
	// or there is no previously-provided http.Handler.
	//
	// The provided http.Handler may be an infuse.Handler itself. However,
	// calls to infuse.Next inside of a nested infuse.Handler are local to that
	// infuse.Handler and will not cause handlers outside of it to be called.
	// Therefore, for infuse.Handlers A and B and http.Handler C,
	//   A.Handle(B).Handle(C)
	// will never serve http.Handler C, but
	//   A.Handle(B.Handle(C))
	// may serve http.Handler C.
	Handle(handler http.Handler) Handler

	// HandleFunc has the same behavior as Handle, but it takes a handler
	// function instead of an http.Handler.
	HandleFunc(handler func(http.ResponseWriter, *http.Request)) Handler

	// Stack has the same behavior as Handle but will implicitly serve an
	// http.Handler attached after the provided http.Handler, if another
	// http.Handler is attached. The provided http.Handler will still only be
	// called if the previously-provided http.Handler calls infuse.Next or
	// there is no previously-provided http.Handler.
	//
	// Stack is useful for serving generic handlers that do not know about
	// infuse. It can also be used to serve another infuse.Handler.
	//
	// Note that for infuse.Handlers A and B and http.Handler C,
	//   A.Handle(B).Handle(C)
	// will never serve http.Handler C. While
	//   A.Stack(B).Handle(C)
	// will always serve http.Handler C after infuse.Handler B is served.
	// However, it is preferable to use
	//   A.Handle(B.Handle(C))
	// as it permits serving C when the last http.Handler attached to B calls
	// infuse.Next.
	//
	// Finally, do not call infuse.Next in a stacked http.Handler unless you
	// intend to serve the rest of the middleware chain multiple times.
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

	sharedResponse := &contextualResponse{ResponseWriter: response}

	current := l
	for ; current.prev != nil; current = current.prev {
		sharedResponse.layers = append(sharedResponse.layers, current)
	}
	current.handler.ServeHTTP(sharedResponse, request)
}

type contextualResponse struct {
	http.ResponseWriter
	context interface{}
	layers  []*layer
}

func (c *contextualResponse) next(request *http.Request) bool {
	if len(c.layers) == 0 {
		return false
	}

	originalLayers := c.layers
	nextLayer := c.layers[len(c.layers)-1]
	c.layers = c.layers[:len(c.layers)-1]
	defer func() { c.layers = originalLayers }()

	nextLayer.handler.ServeHTTP(c, request)
	return true
}

// Next serves the next http.Handler. Next should only be called from an
// http.Handler attached to an infuse.Handler. See infuse.Handler for more
// information.
//
// Next will return false if no subsequent http.Handler is attached, or if the
// provided http.ResponseWriter was not served by an infuse.Handler. For Next
// to successfully serve the correct handler and return true, the response
// must be the same http.ResponseWriter that the current http.Handler was
// called with (via its infuse.Handler).
//
// Calling Next multiple times in the same handler will run the rest of the
// middleware chain attached to the infuse.Handler for each call.
func Next(response http.ResponseWriter, request *http.Request) bool {
	sharedResponse, ok := response.(*contextualResponse)
	return ok && sharedResponse.next(request)
}

// Get will retrieve a context value that is shared by the each http.Handler
// attached to the same infuse.Handler within a single request-response cycle.
// http.Handlers that are attached to different infuse.Handlers do not share
// the same context value.
//
// If the context value is a pointer, map, or slice, changes to the referent
// data will be seen by other http.Handlers that share the context value.
//
// To retrieve a value of a particular type, consider wrapping infuse.Get:
//   func GetMyContext(response http.ResponseWriter) *MyContext {
//      return infuse.Get(response).(*MyContext)
//   }
//
// Example using a map:
//   func HandlerMap(response http.ResponseWriter) map[string]string {
//      handlerMap, ok := infuse.Get(response).(map[string]string)
//      if !ok || handlerMap == nil {
//         handlerMap = make(map[string]string)
//         infuse.Set(response, handlerMap)
//      }
//      return handlerMap
//   }
func Get(response http.ResponseWriter) interface{} {
	return response.(*contextualResponse).context
}

// Set will store a context value that is shared by the each http.Handler
// attached to the same infuse.Handler within a single request-response cycle.
// http.Handlers that are attached to different infuse.Handlers do not share
// the same context value.
func Set(response http.ResponseWriter, value interface{}) {
	response.(*contextualResponse).context = value
}
