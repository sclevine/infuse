package infuse_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sclevine/infuse"
)

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

func createMapHandler(response http.ResponseWriter, request *http.Request) {
	if ok := infuse.Set(response, make(map[string]string)); !ok {
		panic("Failed to set map.")
	}
	infuse.Next(response, request)
}

func buildOutputMapHandler(key string) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		shared := infuse.Get(response).(map[string]string)
		fmt.Fprintf(response, "%s: %s\n", key, shared[key])
		infuse.Next(response, request)
	}
}

func buildSetMapHandler(key, value string) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		shared := infuse.Get(response).(map[string]string)
		shared[key] = value
		infuse.Next(response, request)
	}
}

func serve(t *testing.T, handler infuse.Handler) string {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, &http.Request{})
	return response.Body.String()
}

func testHandlerResponse(t *testing.T, handler infuse.Handler, fixture string) {
	if body := serve(t, handler); body != fixture {
		t.Fatalf("Expected:\n%s\nGot:\n%s\n", fixture, body)
	}
}

func TestNew(t *testing.T) {
	testHandlerResponse(t, infuse.New(), "")
}

func TestHandler(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("first", 1))
	handler = handler.HandleFunc(buildHandler("second", 1))
	handler = handler.HandleFunc(buildHandler("third", 1))
	testHandlerResponse(t, handler, threeHandlerFixture)
}

func TestNestedHandlers(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("third", 1))
	handler = infuse.New().HandleFunc(buildHandler("second", 1)).Handle(handler)
	handler = infuse.New().HandleFunc(buildHandler("first", 1)).Handle(handler)
	testHandlerResponse(t, handler, threeHandlerFixture)
}

func TestComplexNestedHandlers(t *testing.T) {
	firstGroup := infuse.New().HandleFunc(buildHandler("first-first", 1))
	firstGroup = firstGroup.HandleFunc(buildHandler("first-second", 1))

	secondGroup := infuse.New().HandleFunc(buildHandler("second-first", 1))
	secondGroup = secondGroup.HandleFunc(buildHandler("second-second", 1))

	handler := firstGroup.Handle(secondGroup)
	testHandlerResponse(t, handler, complexHandlerFixture)
}

func TestBranchedHandlers(t *testing.T) {
	base := infuse.New().HandleFunc(buildHandler("base", 1))
	first := base.HandleFunc(buildHandler("first", 1))
	second := base.HandleFunc(buildHandler("second", 1))
	handler := first.Handle(second)
	testHandlerResponse(t, handler, branchedHandlerFixture)
}

func TestMultipleNextCalls(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("first", 2))
	handler = handler.HandleFunc(buildHandler("second", 3))
	testHandlerResponse(t, handler, multipleCallHandlerFixture)
}

func TestStackedHandlers(t *testing.T) {
	firstGroup := infuse.New().StackFunc(buildHandler("first stacked", 0))
	firstGroup = firstGroup.StackFunc(buildHandler("second stacked", 0))

	secondGroup := infuse.New().StackFunc(buildHandler("third stacked", 0))
	secondGroup = secondGroup.HandleFunc(buildHandler("fourth handled", 1))
	secondGroup = secondGroup.StackFunc(buildHandler("fifth stacked", 0))

	handler := firstGroup.Stack(secondGroup)
	testHandlerResponse(t, handler, stackedHandlerFixture)
}

func TestGetAndSet(t *testing.T) {
	handler := infuse.New().HandleFunc(createMapHandler)
	handler = handler.HandleFunc(buildSetMapHandler("first key", "first value"))
	handler = handler.HandleFunc(buildSetMapHandler("second key", "second value"))
	handler = handler.HandleFunc(buildSetMapHandler("first key", "new first value"))
	handler = handler.HandleFunc(buildOutputMapHandler("first key"))
	handler = handler.HandleFunc(buildOutputMapHandler("second key"))
	testHandlerResponse(t, handler, "first key: new first value\nsecond key: second value\n")
}

func TestGetAndSetForNestedHandlers(t *testing.T) {
	firstGroup := infuse.New().HandleFunc(createMapHandler)
	firstGroup = firstGroup.HandleFunc(buildSetMapHandler("key", "first value"))

	secondGroup := infuse.New().HandleFunc(createMapHandler)
	secondGroup = secondGroup.HandleFunc(buildSetMapHandler("key", "second value"))
	secondGroup = secondGroup.HandleFunc(buildOutputMapHandler("key"))

	handler := firstGroup.Stack(secondGroup)
	handler = handler.HandleFunc(buildOutputMapHandler("key"))

	testHandlerResponse(t, handler, "key: second value\nkey: first value\n")
}

func TestPanicRecovery(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("first", 1))
	handler = handler.HandleFunc(recoverHandler)
	handler = handler.HandleFunc(buildHandler("second", 1))
	handler = handler.HandleFunc(buildHandler("third", 1))
	handler = handler.HandleFunc(panicHandler)
	testHandlerResponse(t, handler, panicRecoveryHandlerFixture)
}

func TestInvalidResponse(t *testing.T) {
	if context := infuse.Get(nil); context != nil {
		t.Fatalf("Expected nil context from invalid response, got %s.", context)
	}
	if ok := infuse.Set(nil, "value"); ok {
		t.Fatal("Expected failure to set context on invalid response.")
	}
	if ok := infuse.Next(nil, &http.Request{}); ok {
		t.Fatal("Expected failure to serve next handler with invalid response.")
	}
}
