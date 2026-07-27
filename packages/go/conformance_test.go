package aclic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Runs the conformance vectors through the Go binding.
//
// The same files the Rust, Java, and Python runners read. A case this binding answers
// differently is a license that behaves one way on a Go deployment and another way elsewhere,
// which is exactly what one core exists to make impossible.

type vectorCase struct {
	Name   string          `json:"name"`
	Why    string          `json:"why"`
	Input  json.RawMessage `json:"input"`
	Expect json.RawMessage `json:"expect"`
}

type vectorFile struct {
	Kind  string       `json:"kind"`
	About string       `json:"about"`
	Cases []vectorCase `json:"cases"`
}

// load reads one vector file.
func load(t *testing.T, name string) vectorFile {
	t.Helper()

	directory := os.Getenv("ACLIC_CONFORMANCE_DIR")
	if directory == "" {
		directory = filepath.Join("..", "..", "conformance")
	}
	content, err := os.ReadFile(filepath.Join(directory, "vectors", name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}

	var file vectorFile
	if err := json.Unmarshal(content, &file); err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	if len(file.Cases) == 0 {
		t.Fatalf("%s carries no cases", name)
	}
	return file
}

func TestReleaseVectorsHold(t *testing.T) {
	for _, testCase := range load(t, "release.json").Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			var input map[string]any
			if err := json.Unmarshal(testCase.Input, &input); err != nil {
				t.Fatal(err)
			}
			var expect map[string]any
			if err := json.Unmarshal(testCase.Expect, &expect); err != nil {
				t.Fatal(err)
			}

			var answer map[string]any
			if err := Shared.Release(input, &answer); err != nil {
				t.Fatalf("%s: %v", testCase.Why, err)
			}

			switch input["operation"] {
			case "COMPARE":
				actual := sign(int(answer["result"].(float64)))
				if actual != int(expect["result"].(float64)) {
					t.Fatalf("%s: got %d, want %v", testCase.Why, actual, expect["result"])
				}
			default:
				if !equalJSON(answer["result"], expect["result"]) {
					t.Fatalf("%s: got %v, want %v", testCase.Why, answer["result"], expect["result"])
				}
			}
		})
	}
}

func TestFingerprintVectorsHold(t *testing.T) {
	for _, testCase := range load(t, "fingerprint.json").Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			var input map[string]any
			if err := json.Unmarshal(testCase.Input, &input); err != nil {
				t.Fatal(err)
			}

			var answer map[string]any
			if err := Shared.FingerprintCompute(input, &answer); err != nil {
				t.Fatalf("%s: %v", testCase.Why, err)
			}

			var expect map[string]any
			if err := json.Unmarshal(testCase.Expect, &expect); err != nil {
				t.Fatal(err)
			}
			for field, want := range expect {
				if !equalJSON(answer[field], want) {
					t.Fatalf("%s: %s got %v, want %v", testCase.Why, field, answer[field], want)
				}
			}
		})
	}
}

func TestSignatureVectorsHold(t *testing.T) {
	for _, testCase := range load(t, "signature.json").Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			var input map[string]any
			if err := json.Unmarshal(testCase.Input, &input); err != nil {
				t.Fatal(err)
			}

			var answer map[string]any
			if err := Shared.VerifySignature(input, &answer); err != nil {
				t.Fatalf("%s: %v", testCase.Why, err)
			}

			var expect map[string]any
			if err := json.Unmarshal(testCase.Expect, &expect); err != nil {
				t.Fatal(err)
			}
			for field, want := range expect {
				if !equalJSON(answer[field], want) {
					t.Fatalf("%s: %s got %v, want %v", testCase.Why, field, answer[field], want)
				}
			}
		})
	}
}

func TestEvaluationVectorsHold(t *testing.T) {
	for _, testCase := range load(t, "evaluation.json").Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			var input map[string]any
			if err := json.Unmarshal(testCase.Input, &input); err != nil {
				t.Fatal(err)
			}

			var evaluation Evaluation
			if err := Shared.Evaluate(input, &evaluation); err != nil {
				t.Fatalf("%s: %v", testCase.Why, err)
			}

			var expect struct {
				Status            string   `json:"status"`
				Usable            bool     `json:"usable"`
				ErrorCode         string   `json:"errorCode"`
				AvailableFeatures []string `json:"availableFeatures"`
			}
			if err := json.Unmarshal(testCase.Expect, &expect); err != nil {
				t.Fatal(err)
			}

			if evaluation.Status != expect.Status {
				t.Fatalf("%s: status %s, want %s", testCase.Why, evaluation.Status, expect.Status)
			}
			if evaluation.Usable != expect.Usable {
				t.Fatalf("%s: usable %v, want %v", testCase.Why, evaluation.Usable, expect.Usable)
			}
			if evaluation.ErrorCode != expect.ErrorCode {
				t.Fatalf("%s: errorCode %s, want %s", testCase.Why,
					evaluation.ErrorCode, expect.ErrorCode)
			}
			if len(evaluation.AvailableFeatures) != len(expect.AvailableFeatures) {
				t.Fatalf("%s: features %v, want %v", testCase.Why,
					evaluation.AvailableFeatures, expect.AvailableFeatures)
			}
			for index, feature := range expect.AvailableFeatures {
				if evaluation.AvailableFeatures[index] != feature {
					t.Fatalf("%s: features %v, want %v", testCase.Why,
						evaluation.AvailableFeatures, expect.AvailableFeatures)
				}
			}
		})
	}
}

func TestTheGateRefusesBeforeAnythingIsJudged(t *testing.T) {
	err := NewGate().RequireUsable()
	denied, ok := err.(*LicenseDenied)
	if !ok {
		t.Fatalf("expected a refusal, got %v", err)
	}
	if denied.ErrorCode != "LICENSE_NOT_ACTIVATED" {
		t.Fatalf("errorCode %s", denied.ErrorCode)
	}
	if denied.HTTPStatus != Forbidden {
		t.Fatalf("status %d", denied.HTTPStatus)
	}
}

func TestAQuotaRefusesTheOneThatWouldPassTheCeiling(t *testing.T) {
	gate := NewGate()
	gate.Refresh(&Evaluation{
		Status:            "VALID",
		Usable:            true,
		ErrorCode:         "LICENSE_VALID",
		AvailableFeatures: []string{"ATTENDANCE"},
		Limits:            map[string]int64{"USERS": 2},
	})

	if err := gate.CheckQuota("USERS", 1); err != nil {
		t.Fatalf("one below the ceiling: %v", err)
	}
	if err := gate.CheckQuota("USERS", 2); err == nil {
		t.Fatal("at the ceiling, one more should be refused")
	}
	// A code the license does not name is unlimited.
	if err := gate.CheckQuota("READER_PORTS", 1_000_000); err != nil {
		t.Fatalf("an unnamed quota: %v", err)
	}
}

func TestTheMiddlewareRefusesWithTheSharedDocument(t *testing.T) {
	handler := LicenseGateMiddleware(NewGate(), []string{"/license"})(
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/doors", nil))

	if recorder.Code != Forbidden {
		t.Fatalf("status %d", recorder.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["errorCode"] != "LICENSE_NOT_ACTIVATED" || body["status"] != "NOT_ACTIVATED" {
		t.Fatalf("body %v", body)
	}
}

func TestTheLicenseScreenStaysReachableWhileUnlicensed(t *testing.T) {
	served := false
	handler := LicenseGateMiddleware(NewGate(), []string{"/license"})(
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

func TestTheBindingAndTheCoreReadOneSchema(t *testing.T) {
	if Shared.SchemaVersion() != 1 {
		t.Fatalf("schema %d", Shared.SchemaVersion())
	}
	// Against what the workspace declares rather than a literal: the question is whether this
	// binding linked the core built beside it, and a literal would only ask whether someone
	// remembered to edit this line.
	declared := workspaceVersion(t)
	if Shared.Version() != declared {
		t.Fatalf("the core reports %s, the workspace declares %s", Shared.Version(), declared)
	}
}

// workspaceVersion is the version the Rust workspace declares for every crate in it.
func workspaceVersion(t *testing.T) string {
	t.Helper()
	manifest, err := os.ReadFile(filepath.Join("..", "..", "Cargo.toml"))
	if err != nil {
		t.Fatalf("reading the workspace manifest: %v", err)
	}
	for _, line := range strings.Split(string(manifest), "\n") {
		if strings.HasPrefix(line, "version = \"") {
			return strings.Split(line, "\"")[1]
		}
	}
	t.Fatal("the workspace manifest declares no version")
	return ""
}

// sign reduces a comparison to -1, 0, or 1.
func sign(value int) int {
	switch {
	case value < 0:
		return -1
	case value > 0:
		return 1
	default:
		return 0
	}
}

// equalJSON compares two values read out of JSON.
func equalJSON(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return string(leftJSON) == string(rightJSON)
}
