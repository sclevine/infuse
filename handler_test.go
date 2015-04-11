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

func createMapHandler(response http.ResponseWriter, request *http.Request) {
	infuse.Set(response, make(map[string]string))
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
	request, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal("Failed to generate request.")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response.Body.String()
}

func testHandlerResponse(t *testing.T, handler infuse.Handler, fixture string) {
	if body := serve(t, handler); body != fixture {
		t.Fatalf("Expected:\n%s\n\nGot:\n%s", fixture, body)
	}
}

func TestNew(t *testing.T) {
	handler := infuse.New()
	if serve(t, handler) != "" {
		t.Fatal("Failed to serve empty infuse.Handler.")
	}
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
	first := infuse.New().HandleFunc(buildHandler("first-first", 1))
	first = first.HandleFunc(buildHandler("first-second", 1))

	second := infuse.New().HandleFunc(buildHandler("second-first", 1))
	second = second.HandleFunc(buildHandler("second-second", 1))

	handler := first.Handle(second)
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

func TestGetAndSet(t *testing.T) {
	handler := infuse.New().HandleFunc(createMapHandler)
	handler = handler.HandleFunc(buildSetMapHandler("first key", "first value"))
	handler = handler.HandleFunc(buildSetMapHandler("second key", "second value"))
	handler = handler.HandleFunc(buildSetMapHandler("first key", "new first value"))
	handler = handler.HandleFunc(buildOutputMapHandler("first key"))
	handler = handler.HandleFunc(buildOutputMapHandler("second key"))
	testHandlerResponse(t, handler, "first key: new first value\nsecond key: second value\n")
}

}
