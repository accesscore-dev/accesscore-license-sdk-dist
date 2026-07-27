package aclic

import (
	"errors"
	"net/http"
	"strings"
)

// Framework adapters.
//
// An adapter calls the neutral gate and moves what it answered into that framework's response.
// It holds no judgement of its own — not the status code, not the error code, not the message.
// That is why adding a framework stays cheap, and why every framework refuses the same request
// the same way.
//
// gin, echo, and fiber are reached through their own handler shapes without importing them:
// each takes a writer and a request, or something that yields one, and the whole of what an
// adapter does is write the refusal. A product wires the one it uses in three lines rather than
// taking a dependency on all four.

// PathIsOpen reports whether a path stays reachable while the deployment is unlicensed.
//
// Which paths those are is the product's decision: a deployment has to reach its own license
// screen and its own activation endpoints in order to stop being unlicensed, and only the
// product knows where they live.
func PathIsOpen(path string, openPrefixes []string) bool {
	for _, prefix := range openPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// WriteDenial writes a refusal to any standard response writer.
//
// This is the whole of what a framework adapter does, which is why one function serves gin,
// echo, fiber's net/http mode, and the standard library alike.
func WriteDenial(writer http.ResponseWriter, err error) bool {
	var denied *LicenseDenied
	if !errors.As(err, &denied) {
		return false
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(denied.HTTPStatus)
	// A short write here would leave the client with a status and no reason; there is nothing
	// further this can do about it, and the status already carries the refusal.
	_, _ = writer.Write(denied.Body())
	return true
}

// LicenseGateMiddleware closes a net/http deployment while its license does not let it serve.
func LicenseGateMiddleware(gate *Gate, openPrefixes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if PathIsOpen(request.URL.Path, openPrefixes) {
				next.ServeHTTP(writer, request)
				return
			}
			if err := gate.RequireUsable(); err != nil {
				WriteDenial(writer, err)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}

// RequireFeatureMiddleware refuses a request that needs a feature the license does not grant.
func RequireFeatureMiddleware(gate *Gate, feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if err := gate.RequireFeature(feature); err != nil {
				WriteDenial(writer, err)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}

// GinHandler adapts the gate to gin's handler shape.
//
// Taking the context as an interface keeps gin off this module's dependency list while still
// giving a gin application a one-line wiring: gin's *gin.Context satisfies it.
type GinContext interface {
	// Writer is the response writer gin holds.
	Writer() http.ResponseWriter
	// Request is the request being served.
	Request() *http.Request
	// Abort stops the remaining handlers.
	Abort()
	// Next runs the remaining handlers.
	Next()
}

// GinLicenseGate closes a gin deployment while its license does not let it serve.
func GinLicenseGate(gate *Gate, openPrefixes []string) func(GinContext) {
	return func(context GinContext) {
		if PathIsOpen(context.Request().URL.Path, openPrefixes) {
			context.Next()
			return
		}
		if err := gate.RequireUsable(); err != nil {
			WriteDenial(context.Writer(), err)
			context.Abort()
			return
		}
		context.Next()
	}
}

// EchoContext is the part of echo's context an adapter needs.
type EchoContext interface {
	// Response is the response writer echo holds.
	Response() http.ResponseWriter
	// Request is the request being served.
	Request() *http.Request
}

// EchoLicenseGate closes an echo deployment while its license does not let it serve.
//
// It returns the refusal rather than writing it when echo's own error handler is in use, so a
// product keeps one place where responses are shaped.
func EchoLicenseGate(gate *Gate, openPrefixes []string) func(EchoContext) error {
	return func(context EchoContext) error {
		if PathIsOpen(context.Request().URL.Path, openPrefixes) {
			return nil
		}
		return gate.RequireUsable()
	}
}

// FiberContext is the part of fiber's context an adapter needs.
type FiberContext interface {
	// Path is the path being served.
	Path() string
	// Status sets the response status.
	Status(code int) FiberContext
	// Send writes the body.
	Send(body []byte) error
	// Next runs the remaining handlers.
	Next() error
}

// FiberLicenseGate closes a fiber deployment while its license does not let it serve.
func FiberLicenseGate(gate *Gate, openPrefixes []string) func(FiberContext) error {
	return func(context FiberContext) error {
		if PathIsOpen(context.Path(), openPrefixes) {
			return context.Next()
		}
		err := gate.RequireUsable()
		if err == nil {
			return context.Next()
		}

		var denied *LicenseDenied
		if !errors.As(err, &denied) {
			return err
		}
		return context.Status(denied.HTTPStatus).Send(denied.Body())
	}
}
