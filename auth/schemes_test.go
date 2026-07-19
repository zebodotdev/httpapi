package auth

import "testing"

func TestNormalizeAuthorizationSchemes(t *testing.T) {
	schemes := NormalizeAuthorizationSchemes(AuthorizationSchemes{
		ServiceAliases: []string{" Service ", "Assertion", "assertion", ""},
	})

	if schemes.Bearer != DefaultBearerAuthorizationScheme {
		t.Fatalf("bearer = %q, want default", schemes.Bearer)
	}
	if schemes.Service != DefaultServiceAuthorizationScheme {
		t.Fatalf("service = %q, want default", schemes.Service)
	}

	got := schemes.ServiceAuthorizationSchemes()
	want := []string{DefaultServiceAuthorizationScheme, "Assertion"}
	if len(got) != len(want) {
		t.Fatalf("service schemes = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("service schemes = %#v, want %#v", got, want)
		}
	}
}
