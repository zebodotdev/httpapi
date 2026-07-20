// Package gcpapigateway transcribes httpapi routes into the Swagger 2.0 shape
// consumed by GCP API Gateway.
//
// It is a target-specific writer. Endpoint.RouteSpec stays provider-neutral,
// and this package translates route backends, path translation, deadlines,
// authorization metadata, internal markers, and priority into the fields and
// x-extensions expected by GCP API Gateway.
package gcpapigateway
