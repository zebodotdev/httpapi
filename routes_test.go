package httpapi

import "testing"

func TestRoutesFromGroupResolvesMetadataAndPaths(t *testing.T) {
	group := EndpointGroup{
		PathPrefix: "ops",
		Internal:   true,
		Priority:   EndpointPriorityHigh,
	}
	group.RequireAuthorization(AuthorizationKindService)
	group.Add(NewEndpoint(POST, "", noopRouteHandler))
	group.Add(NewEndpoint(GET, "/lookup", noopRouteHandler))

	routes, err := RoutesFromGroup(group)
	if err != nil {
		t.Fatalf("RoutesFromGroup() error = %v", err)
	}
	if len(routes) != 2 {
		t.Fatalf("route count = %d, want 2", len(routes))
	}

	if routes[0].Path != "/ops" {
		t.Fatalf("first path = %q, want /ops", routes[0].Path)
	}
	if routes[0].Method != POST {
		t.Fatalf("first method = %q, want %q", routes[0].Method, POST)
	}
	if !routes[0].Endpoint.IsInternal() {
		t.Fatal("group metadata did not mark endpoint internal")
	}
	if routes[0].Endpoint.Authorization().Kind != AuthorizationKindService {
		t.Fatalf("authorization kind = %q", routes[0].Endpoint.Authorization().Kind)
	}
	if routes[0].Endpoint.Priority() != EndpointPriorityHigh {
		t.Fatalf("priority = %q, want %q", routes[0].Endpoint.Priority(), EndpointPriorityHigh)
	}

	if routes[1].Path != "/ops/lookup" {
		t.Fatalf("second path = %q, want /ops/lookup", routes[1].Path)
	}
	if routes[1].Method != GET {
		t.Fatalf("second method = %q, want %q", routes[1].Method, GET)
	}
}

func TestRoutesWithPathPrefix(t *testing.T) {
	routes, err := RoutesFromEndpoint(NewEndpoint(POST, "/orders/new", noopRouteHandler))
	if err != nil {
		t.Fatalf("RoutesFromEndpoint() error = %v", err)
	}

	routes, err = routes.WithPathPrefix("/v1")
	if err != nil {
		t.Fatalf("WithPathPrefix() error = %v", err)
	}

	if len(routes) != 1 || routes[0].Path != "/v1/orders/new" {
		t.Fatalf("routes = %#v, want /v1/orders/new", routes)
	}
}

func noopRouteHandler(*Req) {}
