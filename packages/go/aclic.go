// Package aclic is the AccessCore license SDK for Go.
//
// The judgement lives in one Rust core that every language binding calls; this package is the
// Go wrapper over it, plus the neutral gate, the host interfaces, and the framework adapters.
//
// Nothing here decides whether a license is valid. That is the core's, and it is the same core
// a Java, Python, or Rust deployment runs — which is what stops two deployments from
// disagreeing about one token.
//
// The core is linked statically, so a Go binary carries it rather than needing a shared library
// installed beside it.
package aclic

/*
#include <stdlib.h>
#include "native/aclic.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

// Return codes the boundary answers with.
const (
	aclicOK              = 0
	aclicNullArgument    = -1
	aclicInvalidUTF8     = -2
	aclicOperationFailed = -3
	aclicPanic           = -4
)

// Error is what a call across the boundary raises when it produces no result.
type Error struct {
	// Operation names the call that failed.
	Operation string
	// Message says what went wrong, carrying the core's error document when there was one.
	Message string
}

func (e *Error) Error() string {
	return e.Operation + ": " + e.Message
}

// reason names a return code for a caller reading a log.
func reason(code C.int32_t) string {
	switch code {
	case aclicNullArgument:
		return "a required argument was null"
	case aclicInvalidUTF8:
		return "the request was not valid UTF-8"
	case aclicOperationFailed:
		return "the operation produced no result"
	case aclicPanic:
		return "the license core failed"
	default:
		return fmt.Sprintf("the license core returned %d", int(code))
	}
}

// Core is the license core, called through the C ABI.
//
// The core holds no state between calls — every judgement is a function of what it was handed —
// so one value serves a whole process and the zero value is ready to use.
type Core struct{}

// Shared is the core every caller uses unless it has a reason not to.
var Shared = &Core{}

// call sends one request across the boundary and reads the answer into out.
func (c *Core) call(
	fn func(*C.char, **C.char) C.int32_t,
	request any,
	out any,
	operation string,
) error {
	encoded, err := json.Marshal(request)
	if err != nil {
		return &Error{Operation: operation, Message: "the request could not be written"}
	}

	cRequest := C.CString(string(encoded))
	defer C.free(unsafe.Pointer(cRequest))

	var answer *C.char
	code := fn(cRequest, &answer)
	return c.read(code, answer, out, operation)
}

// read reads an answer and frees the memory the core handed over.
func (c *Core) read(code C.int32_t, answer *C.char, out any, operation string) error {
	var text string
	if answer != nil {
		text = C.GoString(answer)
		// Freed whatever the outcome: the core allocated it, and a failure path that skipped
		// the free would leak a little on every refusal.
		C.aclic_string_free(answer)
	}

	if code != aclicOK {
		message := reason(code)
		if text != "" {
			message = message + ": " + text
		}
		return &Error{Operation: operation, Message: message}
	}
	if text == "" {
		return &Error{Operation: operation, Message: "the core returned no answer"}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal([]byte(text), out); err != nil {
		return &Error{Operation: operation, Message: "the answer could not be read"}
	}
	return nil
}

// Evaluate judges everything a host knows about its license.
func (c *Core) Evaluate(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_evaluate(r, o)
	}, request, out, "evaluate")
}

// Sign signs a payload into a token, or reads which key pair a key belongs to.
//
// The private key travels in the request as PKCS#8 and is wiped from the core's memory as soon
// as the call is done with it.
func (c *Core) Sign(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_sign(r, o)
	}, request, out, "sign")
}

// VerifySignature checks a token's signature without judging anything else about it.
func (c *Core) VerifySignature(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_verify_signature(r, o)
	}, request, out, "verifySignature")
}

// Fingerprint reads what this machine reports about itself.
func (c *Core) Fingerprint(out any) error {
	var answer *C.char
	code := C.aclic_fingerprint(&answer)
	return c.read(code, answer, out, "fingerprint")
}

// FingerprintCompute hashes machine identifiers, and matches them against what a license names.
//
// Separate from Fingerprint because reading a machine and calculating over what was read are
// different things: one is platform input and output, the other is a pure function the
// conformance vectors can check without a machine to run them on.
func (c *Core) FingerprintCompute(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_fingerprint_compute(r, o)
	}, request, out, "fingerprintCompute")
}

// ActivationRequest writes the activation, heartbeat, or release request a host sends.
func (c *Core) ActivationRequest(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_activation_request(r, o)
	}, request, out, "activationRequest")
}

// ActivationResponse reads the answer an activation server gave.
func (c *Core) ActivationResponse(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_activation_response(r, o)
	}, request, out, "activationResponse")
}

// GenerateKeyPair generates a signing key pair.
func (c *Core) GenerateKeyPair(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_generate_keypair(r, o)
	}, request, out, "generateKeyPair")
}

// ProductKey generates and reads customer-facing product keys.
func (c *Core) ProductKey(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_product_key(r, o)
	}, request, out, "productKey")
}

// Release ranks release versions and checks them against a ceiling.
func (c *Core) Release(request any, out any) error {
	return c.call(func(r *C.char, o **C.char) C.int32_t {
		return C.aclic_release(r, o)
	}, request, out, "release")
}

// SchemaVersion is the payload schema number this build reads and writes.
func (c *Core) SchemaVersion() uint32 {
	return uint32(C.aclic_schema_version())
}

// Version is the SDK version the linked core reports.
func (c *Core) Version() string {
	return C.GoString(C.aclic_version())
}
