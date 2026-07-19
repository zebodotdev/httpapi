// Package transcribers converts httpapi endpoint metadata into OpenAPI-shaped
// documents.
//
// The core httpapi package owns endpoint behavior and metadata. This package
// reads that metadata through exported accessors and writes target-specific
// schemas, such as public OpenAPI documents and GCP API Gateway specs.
package transcribers
