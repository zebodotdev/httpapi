// Package response owns response objects, render helpers, body encoding, and
// HTTP response writing for httpapi endpoint runtimes.
//
// Handlers usually call RenderJSON, RenderErr, or RenderStream. Endpoint
// runtimes call WriteResponse after the handler returns so standard headers,
// timing, request IDs, CORS defaults, streaming, and write deadlines are applied
// consistently.
package response
