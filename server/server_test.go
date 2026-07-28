package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zebodotdev/httpapi/endpoint"
	"github.com/zebodotdev/httpapi/response"
)

func TestNewAppliesDefaults(t *testing.T) {
	srv := New(Config{})

	if srv.Addr != ":8080" {
		t.Fatalf("addr = %q, want :8080", srv.Addr)
	}
	if srv.ReadHeaderTimeout != DefaultReadHeaderTimeout {
		t.Fatalf("read header timeout = %s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != DefaultReadTimeout {
		t.Fatalf("read timeout = %s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != DefaultWriteTimeout {
		t.Fatalf("write timeout = %s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != DefaultIdleTimeout {
		t.Fatalf("idle timeout = %s", srv.IdleTimeout)
	}
	if srv.MaxHeaderBytes != DefaultMaxHeaderBytes {
		t.Fatalf("max header bytes = %d", srv.MaxHeaderBytes)
	}
	if srv.Handler == nil {
		t.Fatal("handler is nil")
	}
}

func TestNewUsesConfiguredValues(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	srv := New(Config{
		Host:              "127.0.0.1",
		Port:              "9090",
		Handler:           handler,
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       4 * time.Second,
		MaxHeaderBytes:    4096,
	})

	if srv.Addr != "127.0.0.1:9090" {
		t.Fatalf("addr = %q, want 127.0.0.1:9090", srv.Addr)
	}
	if srv.Handler == nil {
		t.Fatal("handler is nil")
	}
	if srv.ReadHeaderTimeout != time.Second {
		t.Fatalf("read header timeout = %s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != 2*time.Second {
		t.Fatalf("read timeout = %s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 3*time.Second {
		t.Fatalf("write timeout = %s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 4*time.Second {
		t.Fatalf("idle timeout = %s", srv.IdleTimeout)
	}
	if srv.MaxHeaderBytes != 4096 {
		t.Fatalf("max header bytes = %d", srv.MaxHeaderBytes)
	}
}

func TestAddressPrefersExplicitAddr(t *testing.T) {
	addr := Address(Config{
		Addr: ":7000",
		Host: "127.0.0.1",
		Port: "9090",
	})

	if addr != ":7000" {
		t.Fatalf("addr = %q, want :7000", addr)
	}
}

func TestHandlerUsesEndpointMux(t *testing.T) {
	mux := endpoint.NewMux()
	mux.MustMountEndpoints(endpoint.DefineEndpoint(endpoint.EndpointSpec{
		Method: endpoint.GET,
		Path:   "/ready",
		Handler: func(r *endpoint.Req) {
			response.RenderJSON(r, http.StatusOK, map[string]bool{"ready": true})
		},
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	Handler(Config{Mux: mux}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerAppliesMiddlewareInDeclarationOrder(t *testing.T) {
	events := []string{}
	base := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		events = append(events, "handler")
		w.WriteHeader(http.StatusNoContent)
	})

	handler := Handler(Config{
		Handler: base,
		Middleware: []Middleware{
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					events = append(events, "first-before")
					next.ServeHTTP(w, r)
					events = append(events, "first-after")
				})
			},
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					events = append(events, "second-before")
					next.ServeHTTP(w, r)
					events = append(events, "second-after")
				})
			},
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	want := []string{"first-before", "second-before", "handler", "second-after", "first-after"}
	if len(events) != len(want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %#v, want %#v", events, want)
		}
	}
}

func TestPortFromEnv(t *testing.T) {
	t.Setenv("HTTPAPI_TEST_PORT", " 9090 ")

	if port := PortFromEnv("HTTPAPI_TEST_PORT", "8080"); port != "9090" {
		t.Fatalf("port = %q, want 9090", port)
	}
	if port := PortFromEnv("HTTPAPI_MISSING_PORT", "8080"); port != "8080" {
		t.Fatalf("fallback port = %q, want 8080", port)
	}
}
