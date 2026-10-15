package aclic

import (
	"bytes"
	"encoding/json"
	"testing"
)

// The signing half never becomes a string this process cannot clear.
//
// A key marshalled into the request would be a Go string, and a string cannot be overwritten:
// the key would stay readable in the heap until the collector happened to move it. It travels
// as bytes the caller owns instead, which is only worth something if the key is genuinely
// absent from everything else the boundary hands back.

func TestAGeneratedPairHandsItsSigningHalfOverSeparately(t *testing.T) {
	var described map[string]any
	privateKey, err := Shared.GenerateKeyPair(map[string]any{"algorithm": "Ed25519"}, &described)
	if err != nil {
		t.Fatalf("generating a pair: %v", err)
	}

	if described["algorithm"] != "Ed25519" {
		t.Fatalf("algorithm %v", described["algorithm"])
	}
	if _, ok := described["publicKeyPem"].(string); !ok {
		t.Fatal("the pair came back without its verifying half")
	}

	document, err := json.Marshal(described)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(document, []byte("PRIVATE KEY")) ||
		bytes.Contains(document, []byte("privateKey")) {
		t.Fatalf("the signing half appeared in what describes the pair: %s", document)
	}

	if !bytes.Contains(privateKey, []byte("PRIVATE KEY")) {
		t.Fatal("and it did not arrive beside it")
	}
}

func TestAKeySignsAndTheSignatureReadsBack(t *testing.T) {
	var described map[string]any
	privateKey, err := Shared.GenerateKeyPair(map[string]any{"algorithm": "Ed25519"}, &described)
	if err != nil {
		t.Fatalf("generating a pair: %v", err)
	}

	payload := map[string]any{
		"schemaVersion":          Shared.SchemaVersion(),
		"alg":                    "Ed25519",
		"kid":                    "acps-2026-01",
		"licenseId":              "lic_go",
		"channel":                "PRODUCTION",
		"productCode":            "ACCESSCORE_CONFORMANCE",
		"issuedAt":               "2026-07-25T00:00:00Z",
		"gracePeriodDays":        0,
		"machineFingerprints":    []string{"sha256:abc"},
		"features":               []string{},
		"limits":                 map[string]int{},
		"heartbeatRequired":      false,
		"heartbeatIntervalHours":  0,
		"heartbeatToleranceHours": 0,
	}

	var signed map[string]any
	if err := Shared.Sign(map[string]any{"payload": payload}, privateKey, &signed); err != nil {
		t.Fatalf("signing: %v", err)
	}

	var checked map[string]any
	err = Shared.VerifySignature(map[string]any{
		"token": signed["token"],
		"keys": []map[string]any{
			{"kid": "acps-2026-01", "alg": "Ed25519", "publicKey": signed["publicKeyPem"]},
		},
	}, &checked)
	if err != nil {
		t.Fatalf("checking the signature: %v", err)
	}
	if checked["valid"] != true {
		t.Fatalf("the token this key signed did not verify: %v", checked)
	}
}

func TestSigningWithoutAKeyAnswersWhichPairAPublicKeyBelongsTo(t *testing.T) {
	var described map[string]any
	privateKey, err := Shared.GenerateKeyPair(map[string]any{"algorithm": "Ed25519"}, &described)
	if err != nil {
		t.Fatalf("generating a pair: %v", err)
	}

	var fromPublic map[string]any
	err = Shared.Sign(map[string]any{"publicKey": described["publicKeyPem"]}, nil, &fromPublic)
	if err != nil {
		t.Fatalf("reading a public key: %v", err)
	}

	var fromPrivate map[string]any
	if err := Shared.Sign(map[string]any{}, privateKey, &fromPrivate); err != nil {
		t.Fatalf("reading a private key: %v", err)
	}

	if fromPublic["fingerprint"] != fromPrivate["fingerprint"] {
		t.Fatalf("the halves disagree: %v and %v",
			fromPublic["fingerprint"], fromPrivate["fingerprint"])
	}
}

func TestTheCallerCanClearTheBufferItWasGiven(t *testing.T) {
	// The reason the half arrives as bytes at all. A string offers nothing like this.
	var described map[string]any
	privateKey, err := Shared.GenerateKeyPair(map[string]any{"algorithm": "Ed25519"}, &described)
	if err != nil {
		t.Fatalf("generating a pair: %v", err)
	}

	for index := range privateKey {
		privateKey[index] = 0
	}
	if !bytes.Equal(privateKey, make([]byte, len(privateKey))) {
		t.Fatal("the buffer did not clear")
	}
}
