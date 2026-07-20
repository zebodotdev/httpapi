// Package auth contains provider-neutral authentication and authorization
// contracts used by httpapi.
//
// Services provide the actual credential verifier through request.Authenticator.
// This package only defines the shared session, app, authorization requirement,
// scheme, and audit shapes so callers can integrate bearer credentials, service
// assertions, or custom providers without tying the library to one platform.
package auth
