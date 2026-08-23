package barikoi_test

import (
	"errors"
	"testing"

	barikoi "github.com/barikoi/barikoiapis-golang"
)

func TestErrorMessageFormats(t *testing.T) {
	be := &barikoi.BarikoiError{Message: "Invalid or No Registered Key", StatusCode: 401, Code: "no_registered_key"}
	if got, want := be.Error(), "barikoi: Invalid or No Registered Key (status 401, code no_registered_key)"; got != want {
		t.Errorf("BarikoiError.Error() = %q, want %q", got, want)
	}
	beNoCode := &barikoi.BarikoiError{Message: "boom", StatusCode: 500}
	if got, want := beNoCode.Error(), "barikoi: boom (status 500)"; got != want {
		t.Errorf("BarikoiError.Error() = %q, want %q", got, want)
	}
	ve := &barikoi.ValidationError{Field: "radius", Message: "must be between 0.1 and 100 (kilometers)"}
	if got, want := ve.Error(), `barikoi: validation error on field "radius": must be between 0.1 and 100 (kilometers)`; got != want {
		t.Errorf("ValidationError.Error() = %q, want %q", got, want)
	}
	te := &barikoi.TimeoutError{Message: "barikoi: request timed out"}
	if got := te.Error(); got != "barikoi: request timed out" {
		t.Errorf("TimeoutError.Error() = %q", got)
	}
}

func TestErrorPredicates(t *testing.T) {
	be := &barikoi.BarikoiError{StatusCode: 403}
	if !be.IsAuthError() || be.IsRateLimitError() || be.IsServerError() {
		t.Error("403: want auth error only")
	}
	be = &barikoi.BarikoiError{StatusCode: 429}
	if be.IsAuthError() || !be.IsRateLimitError() || be.IsServerError() {
		t.Error("429: want rate-limit error only")
	}
	be = &barikoi.BarikoiError{StatusCode: 503}
	if be.IsAuthError() || be.IsRateLimitError() || !be.IsServerError() {
		t.Error("503: want server error only")
	}
}

func TestSentinelErrors(t *testing.T) {
	var ve *barikoi.ValidationError
	if !errors.As(error(barikoi.ErrInvalidLatitude), &ve) || ve.Field != "latitude" {
		t.Errorf("ErrInvalidLatitude = %v, want latitude ValidationError", barikoi.ErrInvalidLatitude)
	}
	if !errors.As(error(barikoi.ErrInvalidLongitude), &ve) || ve.Field != "longitude" {
		t.Errorf("ErrInvalidLongitude = %v, want longitude ValidationError", barikoi.ErrInvalidLongitude)
	}
}
