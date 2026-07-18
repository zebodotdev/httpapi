package erreur

import "testing"

func TestURLDefaultsToEmptyWhenURLBuilderIsNotConfigured(t *testing.T) {
	if got := URL("custom_error"); got != "" {
		t.Fatalf("URL() = %q, want empty", got)
	}
}

func TestConfigureURLBuilder(t *testing.T) {
	restore := ConfigureURLBuilder(func(doc ErrorDoc) string {
		return "https://example.test/errors/" + doc.Code
	})
	defer restore()

	if got, want := URL("custom_error"), "https://example.test/errors/custom_error"; got != want {
		t.Fatalf("URL() = %q, want %q", got, want)
	}

	if got, want := URLFor("invalid", "bad_request", "invalid_body", "change_request"), "https://example.test/errors/invalid"; got != want {
		t.Fatalf("URLFor() = %q, want %q", got, want)
	}
}
