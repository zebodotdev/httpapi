package route

import (
	"fmt"
	"net/url"
	"strings"

	httpapi "github.com/zebodotdev/httpapi"
)

// Route is a resolved endpoint mount used by OpenAPI transcribers.
type Route struct {
	Method   httpapi.HttpMethod
	Path     string
	Endpoint httpapi.Endpoint
}

// Routes is an ordered list of resolved endpoint mounts.
type Routes []Route

// FromEndpoint returns the route for one endpoint without a group prefix.
func FromEndpoint(endpoint httpapi.Endpoint) (Routes, error) {
	route, err := fromEndpoint("", endpoint)
	if err != nil {
		return nil, err
	}

	return Routes{route}, nil
}

// FromGroup returns resolved endpoint routes for a group.
func FromGroup(group httpapi.EndpointGroup) (Routes, error) {
	endpoints := group.ResolvedEndpoints()
	if len(endpoints) == 0 {
		return nil, nil
	}

	routes := make(Routes, 0, len(endpoints))
	for _, endpoint := range endpoints {
		route, err := fromEndpoint(group.PathPrefix, endpoint)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// FromGroups returns resolved endpoint routes for several groups while
// preserving group and endpoint declaration order.
func FromGroups(groups ...httpapi.EndpointGroup) (Routes, error) {
	var routes Routes
	for _, group := range groups {
		groupRoutes, err := FromGroup(group)
		if err != nil {
			return nil, err
		}
		routes = append(routes, groupRoutes...)
	}

	return routes, nil
}

// WithPathPrefix returns a copy of the routes with prefix applied to every path.
func (routes Routes) WithPathPrefix(prefix string) (Routes, error) {
	if len(routes) == 0 {
		return nil, nil
	}

	prefixed := make(Routes, 0, len(routes))
	for _, route := range routes {
		path, err := JoinPath(prefix, route.Path)
		if err != nil {
			return nil, err
		}
		route.Path = path
		prefixed = append(prefixed, route)
	}

	return prefixed, nil
}

func fromEndpoint(prefix string, endpoint httpapi.Endpoint) (Route, error) {
	path, err := JoinPath(prefix, endpoint.Pattern())
	if err != nil {
		return Route{}, err
	}

	return Route{
		Method:   endpoint.Method(),
		Path:     path,
		Endpoint: endpoint,
	}, nil
}

// JoinPath joins route path fragments using URL path semantics and returns a
// normalized path suitable for OpenAPI transcription.
func JoinPath(prefix, pattern string) (string, error) {
	prefix = strings.TrimSpace(prefix)
	pattern = strings.TrimSpace(pattern)
	if prefix == "" {
		prefix = "/"
	}
	if pattern == "" {
		pattern = "/"
	}

	path, err := url.JoinPath(prefix, pattern)
	if err != nil {
		return "", fmt.Errorf("openapi/internal/route: join path: %w", err)
	}
	path, err = url.PathUnescape(path)
	if err != nil {
		return "", fmt.Errorf("openapi/internal/route: unescape path: %w", err)
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}

	return path, nil
}
