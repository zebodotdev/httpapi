// Package spec contains the shared OpenAPI document model used by target
// transcribers.
//
// It intentionally models the subset of OpenAPI and Swagger used by httpapi
// transcribers. Target packages such as openapi31 and gcpapigateway translate
// endpoint metadata into this shape, including x-extensions where appropriate.
package spec
