package endpoint

import "testing"

func TestNormalizeContentTypes(t *testing.T) {
	got := NormalizeContentTypes("", TextPlain, " "+ApplicationJson+" ", TextPlain)
	want := []ContentType{TextPlain, ApplicationJson}

	if len(got) != len(want) {
		t.Fatalf("content types = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("content types = %#v, want %#v", got, want)
		}
	}
}

func TestValidateContentTypeAcceptsParameters(t *testing.T) {
	if err := ValidateContentType("text/plain; charset=utf-8", []ContentType{TextPlain}); err != nil {
		t.Fatalf("validate content type: %v", err)
	}
}
