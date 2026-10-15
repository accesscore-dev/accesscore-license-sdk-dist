package aclic

import (
	"errors"
	"reflect"
	"testing"
)

// A release above the contract ceiling keeps the deployment usable and grants no feature. The
// refusal has to name the release and the ceiling: the feature is in the license, and a refusal
// that names it sends the operator to the license to look for something already there.
func TestAReleaseAboveTheCeilingIsRefusedForTheReleaseNotTheFeature(t *testing.T) {
	gate := NewGate()
	gate.Refresh(&Evaluation{
		Status:    "RELEASE_NOT_ENTITLED",
		Usable:    true,
		ErrorCode: "LICENSE_RELEASE_NOT_ENTITLED",
		Denial: &Denial{
			MessageKey: "error.license.releaseNotEntitled",
			Arguments:  []any{"2.0.0", "1.1.0"},
		},
		AvailableFeatures: []string{},
	})

	if err := gate.RequireUsable(); err != nil {
		t.Fatalf("a release above the ceiling stopped the deployment: %v", err)
	}

	var denied *LicenseDenied
	if !errors.As(gate.RequireFeature("ATTENDANCE"), &denied) {
		t.Fatal("a release above the ceiling let a feature through")
	}
	if denied.ErrorCode != "LICENSE_RELEASE_NOT_ENTITLED" {
		t.Fatalf("refused with %s", denied.ErrorCode)
	}
	if denied.Status != "RELEASE_NOT_ENTITLED" {
		t.Fatalf("refused under status %s", denied.Status)
	}
	if denied.Denial.MessageKey != "error.license.releaseNotEntitled" {
		t.Fatalf("refused with message key %s", denied.Denial.MessageKey)
	}
	if !reflect.DeepEqual(denied.Denial.Arguments, []any{"2.0.0", "1.1.0"}) {
		t.Fatalf("refused with arguments %v", denied.Denial.Arguments)
	}
	if denied.HTTPStatus != Forbidden {
		t.Fatalf("refused with %d", denied.HTTPStatus)
	}
}
