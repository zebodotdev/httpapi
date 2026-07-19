package response

import (
	"net/http"
	"strings"
	"testing"
)

func TestRenderJSONSetsResponse(t *testing.T) {
	target := &testTarget{}

	RenderJSON(target, http.StatusCreated, map[string]string{"ok": "true"})

	if target.res == nil {
		t.Fatal("response was not set")
	}
	if target.res.Status != http.StatusCreated {
		t.Fatalf("status = %d, want %d", target.res.Status, http.StatusCreated)
	}
	if target.res.ContentType != ApplicationJson {
		t.Fatalf("content type = %q, want %q", target.res.ContentType, ApplicationJson)
	}
}

func TestEncodeResponseBodyEncodesText(t *testing.T) {
	body, err := EncodeResponseBody(&Res{
		ContentType: TextPlain,
		Body:        "hello",
	})

	if err != nil {
		t.Fatalf("EncodeResponseBody error = %v", err)
	}
	if string(body) != "hello" {
		t.Fatalf("body = %q, want hello", body)
	}
}

func TestEncodeResponseBodyEncodesJSON(t *testing.T) {
	body, err := EncodeResponseBody(&Res{
		ContentType: ApplicationJson,
		Body:        map[string]string{"ok": "true"},
	})

	if err != nil {
		t.Fatalf("EncodeResponseBody error = %v", err)
	}
	if !strings.Contains(string(body), `"ok":"true"`) {
		t.Fatalf("body = %q, want json body", body)
	}
}

type testTarget struct {
	res *Res
}

func (t *testTarget) SetResponse(res *Res) {
	t.res = res
}
