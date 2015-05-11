package infuse_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sclevine/infuse"
)

func TestNew(t *testing.T) {
	testHandlerResponse(t, serve(infuse.New()), "")
}

var threeHandlerFixture = `
start first
attempting next for first
start second
attempting next for second
start third
attempting next for third
no next for third
end third
finished next for second
end second
finished next for first
end first`

func TestHandler(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("first", 1))
	handler = handler.HandleFunc(buildHandler("second", 1))
	handler = handler.HandleFunc(buildHandler("third", 1))
	testHandlerResponse(t, serve(handler), threeHandlerFixture)
}

func TestNestedHandlers(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("third", 1))
	handler = infuse.New().HandleFunc(buildHandler("second", 1)).Handle(handler)
	handler = infuse.New().HandleFunc(buildHandler("first", 1)).Handle(handler)
	testHandlerResponse(t, serve(handler), threeHandlerFixture)
}

var complexHandlerFixture = `
start first-first
attempting next for first-first
start first-second
attempting next for first-second
start second-first
attempting next for second-first
start second-second
attempting next for second-second
no next for second-second
end second-second
finished next for second-first
end second-first
finished next for first-second
end first-second
finished next for first-first
end first-first`

func TestComplexNestedHandlers(t *testing.T) {
	firstGroup := infuse.New().HandleFunc(buildHandler("first-first", 1))
	firstGroup = firstGroup.HandleFunc(buildHandler("first-second", 1))

	secondGroup := infuse.New().HandleFunc(buildHandler("second-first", 1))
	secondGroup = secondGroup.HandleFunc(buildHandler("second-second", 1))

	handler := firstGroup.Handle(secondGroup)
	testHandlerResponse(t, serve(handler), complexHandlerFixture)
}

var branchedHandlerFixture = `
start base
attempting next for base
start first
attempting next for first
start base
attempting next for base
start second
attempting next for second
no next for second
end second
finished next for base
end base
finished next for first
end first
finished next for base
end base`

func TestBranchedHandlers(t *testing.T) {
	base := infuse.New().HandleFunc(buildHandler("base", 1))
	first := base.HandleFunc(buildHandler("first", 1))
	second := base.HandleFunc(buildHandler("second", 1))
	handler := first.Handle(second)
	testHandlerResponse(t, serve(handler), branchedHandlerFixture)
}

var multipleCallHandlerFixture = `
start first
attempting next for first
start second
attempting next for second
no next for second
attempting next for second
no next for second
attempting next for second
no next for second
end second
finished next for first
attempting next for first
start second
attempting next for second
no next for second
attempting next for second
no next for second
attempting next for second
no next for second
end second
finished next for first
end first`

func TestMultipleNextCalls(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("first", 2))
	handler = handler.HandleFunc(buildHandler("second", 3))
	testHandlerResponse(t, serve(handler), multipleCallHandlerFixture)
}

var stackedHandlerFixture = `
start first stacked
end first stacked
start second stacked
end second stacked
start third stacked
end third stacked
start fourth handled
attempting next for fourth handled
start fifth stacked
end fifth stacked
finished next for fourth handled
end fourth handled`

func TestStackedHandlers(t *testing.T) {
	firstGroup := infuse.New().StackFunc(buildHandler("first stacked", 0))
	firstGroup = firstGroup.StackFunc(buildHandler("second stacked", 0))

	secondGroup := infuse.New().StackFunc(buildHandler("third stacked", 0))
	secondGroup = secondGroup.HandleFunc(buildHandler("fourth handled", 1))
	secondGroup = secondGroup.StackFunc(buildHandler("fifth stacked", 0))

	handler := firstGroup.Stack(secondGroup)
	testHandlerResponse(t, serve(handler), stackedHandlerFixture)
}

var panicRecoveryHandlerFixture = `
start first
attempting next for first
start recoverable
start second
attempting next for second
start third
attempting next for third
panicking
recovered and attempting next
start second
attempting next for second
start third
attempting next for third
already panicked
finished next for third
end third
finished next for second
end second
finished next after recovery
finished next for first
end first`

func TestPanicRecovery(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("first", 1))
	handler = handler.HandleFunc(recoverHandler)
	handler = handler.HandleFunc(buildHandler("second", 1))
	handler = handler.HandleFunc(buildHandler("third", 1))
	handler = handler.HandleFunc(panicHandler)
	testHandlerResponse(t, serve(handler), panicRecoveryHandlerFixture)
}

func TestInvalidResponseForNext(t *testing.T) {
	if ok := infuse.Next(nil, &http.Request{}); ok {
		t.Fatal("Expected failure to serve next handler with invalid response.")
	}
}

func buildHandler(name string, nexts int) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(response, "start %s\n", name)

		for i := 0; i < nexts; i++ {
			fmt.Fprintf(response, "attempting next for %s\n", name)
			if infuse.Next(response, request) {
				fmt.Fprintf(response, "finished next for %s\n", name)
			} else {
				fmt.Fprintf(response, "no next for %s\n", name)
			}
		}

		fmt.Fprintf(response, "end %s\n", name)
	}
}

func panicHandler(response http.ResponseWriter, _ *http.Request) {
	if infuse.Get(response) != nil {
		fmt.Fprintf(response, "already panicked\n")
	} else {
		fmt.Fprintf(response, "panicking\n")
		infuse.Set(response, true)
		panic("some error")
	}
}

func recoverHandler(response http.ResponseWriter, request *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(response, "recovered and attempting next")
			infuse.Next(response, request)
			fmt.Fprintln(response, "finished next after recovery")
		}
	}()
	fmt.Fprintln(response, "start recoverable")
	infuse.Next(response, request)
	fmt.Fprintln(response, "end recoverable")
}
