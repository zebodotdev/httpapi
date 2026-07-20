package route

import (
	"testing"

	endpointpkg "github.com/zebodotdev/httpapi/endpoint"
)

func TestFromGroupResolvesMetadataAndPaths(t *testing.T) {
	group := endpointpkg.EndpointGroup{
		PathPrefix: "ops",
		Internal:   true,
		Priority:   endpointpkg.EndpointPriorityHigh,
	}
	group.RequireAuthorization(endpointpkg.AuthorizationKindService)
	group.Add(endpointpkg.NewEndpoint(endpointpkg.POST, "", noopRouteHandler))
	group.Add(endpointpkg.NewEndpoint(endpointpkg.GET, "/lookup", noopRouteHandler))

	routes, err := FromGroup(group)
	if err != nil {
		t.Fatalf("FromGroup() error = %v", err)
	}
	if len(routes) != 2 {
		t.Fatalf("route count = %d, want 2", len(routes))
	}

	if routes[0].Path != "/ops" {
		t.Fatalf("first path = %q, want /ops", routes[0].Path)
	}
	if routes[0].Method != endpointpkg.POST {
		t.Fatalf("first method = %q, want %q", routes[0].Method, endpointpkg.POST)
	}
	if !routes[0].Endpoint.IsInternal() {
		t.Fatal("group metadata did not mark endpoint internal")
	}
	if routes[0].Endpoint.Authorization().Kind != endpointpkg.AuthorizationKindService {
		t.Fatalf("authorization kind = %q", routes[0].Endpoint.Authorization().Kind)
	}
	if routes[0].Endpoint.Priority() != endpointpkg.EndpointPriorityHigh {
		t.Fatalf("priority = %q, want %q", routes[0].Endpoint.Priority(), endpointpkg.EndpointPriorityHigh)
	}

	if routes[1].Path != "/ops/lookup" {
		t.Fatalf("second path = %q, want /ops/lookup", routes[1].Path)
	}
	if routes[1].Method != endpointpkg.GET {
		t.Fatalf("second method = %q, want %q", routes[1].Method, endpointpkg.GET)
	}
}

func TestRoutesWithPathPrefix(t *testing.T) {
	routes, err := FromEndpoint(endpointpkg.NewEndpoint(endpointpkg.POST, "/orders/new", noopRouteHandler))
	if err != nil {
		t.Fatalf("FromEndpoint() error = %v", err)
	}

	routes, err = routes.WithPathPrefix("/v1")
	if err != nil {
		t.Fatalf("WithPathPrefix() error = %v", err)
	}

	if len(routes) != 1 || routes[0].Path != "/v1/orders/new" {
		t.Fatalf("routes = %#v, want /v1/orders/new", routes)
	}
}

func noopRouteHandler(*endpointpkg.Req) {}
