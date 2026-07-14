package notification

import (
	"strings"
	"testing"
)

func TestPasswordResetHTMLIncludesRecipientAndCode(t *testing.T) {
	body := passwordResetHTML("普通用户", "123456")
	for _, expected := range []string{"普通用户", "123456", "RentNestHub"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected %q in reset email", expected)
		}
	}
}
