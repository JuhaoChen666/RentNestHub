package httpapi

import (
	"regexp"
	"testing"
)

func TestResetCodeHasSixDigits(t *testing.T) {
	code, err := resetCode()
	if err != nil {
		t.Fatalf("reset code: %v", err)
	}
	if !regexp.MustCompile(`^\d{6}$`).MatchString(code) {
		t.Fatalf("unexpected reset code %q", code)
	}
}
