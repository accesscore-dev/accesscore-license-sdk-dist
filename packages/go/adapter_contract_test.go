package aclic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Every adapter answers the same scenario the same way.
//
// This is what makes adding a framework cheap: an adapter moves a gate decision into a response
// and holds no judgement of its own, so the same refusal has to come back with the same status,
// the same error code, and the same body whichever framework served it. A new adapter is
// finished when it passes these.
//
// The gin, echo, and fiber adapters are reached through their own context interfaces rather than
// their packages, so the contract runs against all four without this module depending on any of
// them.

// unlicensedGate has judged nothing, which is the refusal every adapter is checked on.
func unlicensedGate() *Gate {
	return NewGate()
}

// licensedGate holds a judgement that lets the deployment serve.
func licensedGate() *Gate {
	gate := NewGate()
	gate.Refresh(&Evaluation{
		Status:            "VALID",
		Usable:            true,
		ErrorCode:         "LICENSE_VALID",
		AvailableFeatures: []string{"ATTENDANCE"},
	})
	return gate
}

// assertSharedDocument checks the one document every adapter writes.
func assertSharedDocument(t *testing.T, adapter string, status int, body []byte) {
	t.Helper()

	if status != Forbidden {
		t.Fatalf("%s answered %d, not %d", adapter, status, Forbidden)
	}

	var document map[string]any
	if err := json.Unmarshal(body, &document); err != nil {
		t.Fatalf("%s wrote a body that is not JSON: %v", adapter, err)
	}
	if document["errorCode"] != "LICENSE_NOT_ACTIVATED" {
		t.Fatalf("%s wrote errorCode %v", adapter, document["errorCode"])
	}
	if document["status"] != "NOT_ACTIVATED" {
		t.Fatalf("%s wrote status %v", adapter, document["status"])
	}
	if key, ok := document["messageKey"].(string); !ok || len(key) < len("error.license.") ||
		key[:len("error.license.")] != "error.license." {
		t.Fatalf("%s wrote messageKey %v", adapter, document["messageKey"])
	}
	if _, ok := document["arguments"].([]any); !ok {
		t.Fatalf("%s wrote arguments %v", adapter, document["arguments"])
	}
}

// ---------------------------------------------------------------------------------------
// net/http
// ---------------------------------------------------------------------------------------

func TestNetHTTPRefusesWithTheSharedDocument(t *testing.T) {
	handler := LicenseGateMiddleware(unlicensedGate(), []string{"/license"})(
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/doors", nil))

	assertSharedDocument(t, "net/http", recorder.Code, recorder.Body.Bytes())
}

func TestNetHTTPLetsTheLicenseScreenThroughWhileUnlicensed(t *testing.T) {
	served := false
	handler := LicenseGateMiddleware(unlicensedGate(), []string{"/license"})(
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			served = true
			writer.WriteHeader(http.StatusOK)
		}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/license/status", nil))

	// A deployment has to reach its own license screen in order to stop being unlicensed.
	if !served {
		t.Fatal("the license screen was closed")
	}
}

func TestNetHTTPServesWhenTheLicenseAllowsIt(t *testing.T) {
	served := false
	handler := LicenseGateMiddleware(licensedGate(), nil)(
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			served = true
			writer.WriteHeader(http.StatusOK)
		}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/doors", nil))

	if !served {
		t.Fatal("a license that lets the deployment serve did not")
	}
}

// ---------------------------------------------------------------------------------------
// gin
// ---------------------------------------------------------------------------------------

// ginContext is what gin's *gin.Context gives the adapter, without gin.
type ginContext struct {
	writer   http.ResponseWriter
	request  *http.Request
	aborted  bool
	followed bool
}

func (c *ginContext) Writer() http.ResponseWriter { return c.writer }
func (c *ginContext) Request() *http.Request      { return c.request }
func (c *ginContext) Abort()                      { c.aborted = true }
func (c *ginContext) Next()                       { c.followed = true }

func ginRequest(path string) (*ginContext, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	return &ginContext{
		writer:  recorder,
		request: httptest.NewRequest(http.MethodGet, path, nil),
	}, recorder
}

func TestGinRefusesWithTheSharedDocument(t *testing.T) {
	context, recorder := ginRequest("/api/v1/doors")
	GinLicenseGate(unlicensedGate(), []string{"/license"})(context)

	assertSharedDocument(t, "gin", recorder.Code, recorder.Body.Bytes())
	if !context.aborted {
		t.Fatal("gin refused without stopping the handlers behind it")
	}
}

func TestGinLetsTheLicenseScreenThroughWhileUnlicensed(t *testing.T) {
	context, _ := ginRequest("/license/status")
	GinLicenseGate(unlicensedGate(), []string{"/license"})(context)

	if !context.followed || context.aborted {
		t.Fatal("the license screen was closed")
	}
}

func TestGinServesWhenTheLicenseAllowsIt(t *testing.T) {
	context, _ := ginRequest("/api/v1/doors")
	GinLicenseGate(licensedGate(), nil)(context)

	if !context.followed || context.aborted {
		t.Fatal("a license that lets the deployment serve did not")
	}
}

// ---------------------------------------------------------------------------------------
// echo
// ---------------------------------------------------------------------------------------

// echoContext is what echo's Context gives the adapter, without echo.
type echoContext struct {
	writer  http.ResponseWriter
	request *http.Request
}

func (c *echoContext) Response() http.ResponseWriter { return c.writer }
func (c *echoContext) Request() *http.Request        { return c.request }

func echoRequest(path string) (*echoContext, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	return &echoContext{
		writer:  recorder,
		request: httptest.NewRequest(http.MethodGet, path, nil),
	}, recorder
}

func TestEchoRefusesWithTheSharedDocument(t *testing.T) {
	// echo hands the refusal to its own error handler rather than writing it, so the contract is
	// checked on what that handler receives: the same denial, rendered by the same function every
	// other adapter writes with.
	context, recorder := echoRequest("/api/v1/doors")
	err := EchoLicenseGate(unlicensedGate(), []string{"/license"})(context)
	if err == nil {
		t.Fatal("echo let an unlicensed request through")
	}
	if !WriteDenial(recorder, err) {
		t.Fatal("echo returned something that is not a license refusal")
	}

	assertSharedDocument(t, "echo", recorder.Code, recorder.Body.Bytes())
}

func TestEchoLetsTheLicenseScreenThroughWhileUnlicensed(t *testing.T) {
	context, _ := echoRequest("/license/status")
	if err := EchoLicenseGate(unlicensedGate(), []string{"/license"})(context); err != nil {
		t.Fatalf("the license screen was closed: %v", err)
	}
}

func TestEchoServesWhenTheLicenseAllowsIt(t *testing.T) {
	context, _ := echoRequest("/api/v1/doors")
	if err := EchoLicenseGate(licensedGate(), nil)(context); err != nil {
		t.Fatalf("a license that lets the deployment serve did not: %v", err)
	}
}

// ---------------------------------------------------------------------------------------
// fiber
// ---------------------------------------------------------------------------------------

// fiberContext is what fiber's *fiber.Ctx gives the adapter, without fiber.
type fiberContext struct {
	path     string
	status   int
	body     []byte
	followed bool
}

func (c *fiberContext) Path() string { return c.path }

func (c *fiberContext) Status(code int) FiberContext {
	c.status = code
	return c
}

func (c *fiberContext) Send(body []byte) error {
	c.body = body
	return nil
}

func (c *fiberContext) Next() error {
	c.followed = true
	return nil
}

func TestFiberRefusesWithTheSharedDocument(t *testing.T) {
	context := &fiberContext{path: "/api/v1/doors"}
	if err := FiberLicenseGate(unlicensedGate(), []string{"/license"})(context); err != nil {
		t.Fatalf("fiber returned an error instead of writing the refusal: %v", err)
	}

	assertSharedDocument(t, "fiber", context.status, context.body)
	if context.followed {
		t.Fatal("fiber refused and ran the handlers behind it anyway")
	}
}

func TestFiberLetsTheLicenseScreenThroughWhileUnlicensed(t *testing.T) {
	context := &fiberContext{path: "/license/status"}
	if err := FiberLicenseGate(unlicensedGate(), []string{"/license"})(context); err != nil {
		t.Fatalf("the license screen was closed: %v", err)
	}
	if !context.followed {
		t.Fatal("the license screen was closed")
	}
}

func TestFiberServesWhenTheLicenseAllowsIt(t *testing.T) {
	context := &fiberContext{path: "/api/v1/doors"}
	if err := FiberLicenseGate(licensedGate(), nil)(context); err != nil {
		t.Fatalf("a license that lets the deployment serve did not: %v", err)
	}
	if !context.followed {
		t.Fatal("a license that lets the deployment serve did not")
	}
}
