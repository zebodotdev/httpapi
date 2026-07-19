package httpapi

import (
	"fmt"
	"net/url"
	"strings"
)

// Route is a resolved endpoint mount that can be consumed by documentation,
// gateway, and other read-only route processors.
type Route struct {
	Method   HttpMethod
	Path     string
	Endpoint Endpoint
}

// Routes is an ordered list of resolved endpoint mounts.
type Routes []Route

// RoutesFromEndpoint returns the route for one endpoint without a group prefix.
func RoutesFromEndpoint(endpoint Endpoint) (Routes, error) {
	route, err := routeFromEndpoint("", endpoint)
	if err != nil {
		return nil, err
	}

	return Routes{route}, nil
}

// RoutesFromGroup returns resolved endpoint routes for a group.
func RoutesFromGroup(group EndpointGroup) (Routes, error) {
	endpoints := group.ResolvedEndpoints()
	if len(endpoints) == 0 {
		return nil, nil
	}

	routes := make(Routes, 0, len(endpoints))
	for _, endpoint := range endpoints {
		route, err := routeFromEndpoint(group.PathPrefix, endpoint)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// RoutesFromGroups returns resolved endpoint routes for several groups while
// preserving group and endpoint declaration order.
func RoutesFromGroups(groups ...EndpointGroup) (Routes, error) {
	var routes Routes
	for _, group := range groups {
		groupRoutes, err := RoutesFromGroup(group)
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
		path, err := JoinRoutePath(prefix, route.Path)
		if err != nil {
			return nil, err
		}
		route.Path = path
		prefixed = append(prefixed, route)
	}

	return prefixed, nil
}

func routeFromEndpoint(prefix string, endpoint Endpoint) (Route, error) {
	path, err := JoinRoutePath(prefix, endpoint.Pattern())
	if err != nil {
		return Route{}, err
	}

	return Route{
		Method:   endpoint.Method(),
		Path:     path,
		Endpoint: endpoint,
	}, nil
}

// JoinRoutePath joins route path fragments using URL path semantics and returns
// a normalized path suitable for mounting and transcription.
func JoinRoutePath(prefix, pattern string) (string, error) {
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
		return "", fmt.Errorf("httpapi: join route path: %w", err)
	}
	path, err = url.PathUnescape(path)
	if err != nil {
		return "", fmt.Errorf("httpapi: unescape route path: %w", err)
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
