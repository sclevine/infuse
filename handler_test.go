package infuse_test

import (
	"net/http"
	"testing"

	"github.com/sclevine/infuse"
)

func TestNew(t *testing.T) {
	testHandlerResponse(t, infuse.New(), "")
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
	testHandlerResponse(t, handler, threeHandlerFixture)
}

func TestNestedHandlers(t *testing.T) {
	handler := infuse.New().HandleFunc(buildHandler("third", 1))
	handler = infuse.New().HandleFunc(buildHandler("second", 1)).Handle(handler)
	handler = infuse.New().HandleFunc(buildHandler("first", 1)).Handle(handler)
	testHandlerResponse(t, handler, threeHandlerFixture)
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
	testHandlerResponse(t, handler, complexHandlerFixture)
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
	testHandlerResponse(t, handler, branchedHandlerFixture)
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
	testHandlerResponse(t, handler, multipleCallHandlerFixture)
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
