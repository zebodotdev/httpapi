package transcribers

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	httpapi "github.com/zebodotdev/httpapi"
	authpkg "github.com/zebodotdev/httpapi/auth"
	endpointpkg "github.com/zebodotdev/httpapi/endpoint"
)

type ContentType = endpointpkg.ContentType
type HttpMethod = endpointpkg.Method
type RouteBackend = endpointpkg.RouteBackend
type RoutePathMode = endpointpkg.RoutePathMode

const (
	GET             HttpMethod  = endpointpkg.GET
	POST            HttpMethod  = endpointpkg.POST
	ApplicationJson ContentType = endpointpkg.ApplicationJson

	RoutePathModeAppend   RoutePathMode = endpointpkg.RoutePathModeAppend
	RoutePathModeConstant RoutePathMode = endpointpkg.RoutePathModeConstant

	defaultOpenAPIDocumentSpecVersion    = "3.1.1"
	defaultGCPGatewayDocumentSpecVersion = "2.0"
	defaultOpenAPIDocumentTitle          = "http api"
	defaultOpenAPIDocumentDescription    = "machine-readable representation of the http api"
	defaultGCPGatewayScheme              = "https"
	gcpGatewayPathTranslationAppend      = "APPEND_PATH_TO_ADDRESS"
	gcpGatewayPathTranslationConstant    = "CONSTANT_ADDRESS"
	gcpGatewayBackendDeadlineMax         = 600 * time.Second
	placeholderResponseDescription       = "Required placeholder response."
)

var (
	ErrInternalEndpointPublicOpenAPI = errors.New(
		"httpapi: internal endpoint cannot be transcribed in public openapi mode",
	)
	ErrOpenAPIDocumentVersionRequired = errors.New(
		"httpapi: openapi document version is required",
	)
	ErrGCPGatewayBackendAddressRequired = errors.New(
		"httpapi: gcp api gateway backend address is required",
	)
	ErrGCPGatewayBackendDeadlineExceeded = errors.New(
		"httpapi: gcp api gateway backend deadline exceeds maximum",
	)

	ErrGatewayBackendAddressRequired  = ErrGCPGatewayBackendAddressRequired
	ErrGatewayBackendDeadlineExceeded = ErrGCPGatewayBackendDeadlineExceeded
)

// OpenAPITranscriptionMode selects the OpenAPI surface being generated.
type OpenAPITranscriptionMode string

const (
	// OpenAPITranscriptionModePublic emits public OpenAPI path entries.
	OpenAPITranscriptionModePublic OpenAPITranscriptionMode = "public"
	// OpenAPITranscriptionModeGateway emits the default gateway path entries.
	//
	// Deprecated: use OpenAPITranscriptionModeGCPGateway when generating GCP API
	// Gateway specs. Future gateway targets should define explicit modes.
	OpenAPITranscriptionModeGateway OpenAPITranscriptionMode = OpenAPITranscriptionModeGCPGateway
	// OpenAPITranscriptionModeGCPGateway emits GCP API Gateway path entries.
	OpenAPITranscriptionModeGCPGateway OpenAPITranscriptionMode = "gcp_api_gateway"
)

// OpenAPITranscriptionOption configures OpenAPI transcription.
type OpenAPITranscriptionOption func(*openAPITranscriptionConfig)

type openAPITranscriptionConfig struct {
	PathPrefix            string
	GatewayBackendAddress string
	GatewayHost           string
	DocumentTitle         string
	DocumentDescription   string
	DocumentVersion       string
	PublicServerURL       string
}

// WithOpenAPIPathPrefix prefixes all generated path keys.
func WithOpenAPIPathPrefix(prefix string) OpenAPITranscriptionOption {
	return func(cfg *openAPITranscriptionConfig) {
		cfg.PathPrefix = prefix
	}
}

// WithGCPGatewayBackendAddress sets the Cloud Run/backend address for gateway mode.
func WithGCPGatewayBackendAddress(address string) OpenAPITranscriptionOption {
	return WithGatewayBackendAddress(address)
}

// WithGatewayBackendAddress sets the backend address for gateway mode.
func WithGatewayBackendAddress(address string) OpenAPITranscriptionOption {
	return func(cfg *openAPITranscriptionConfig) {
		cfg.GatewayBackendAddress = strings.TrimSpace(address)
	}
}

// WithGCPGatewayHost sets the API Gateway host used by gateway documents.
func WithGCPGatewayHost(host string) OpenAPITranscriptionOption {
	return WithGatewayHost(host)
}

// WithGatewayHost sets the API host used by gateway documents.
func WithGatewayHost(host string) OpenAPITranscriptionOption {
	return func(cfg *openAPITranscriptionConfig) {
		cfg.GatewayHost = strings.TrimSpace(host)
	}
}

// WithOpenAPITitle sets the generated document title.
func WithOpenAPITitle(title string) OpenAPITranscriptionOption {
	return func(cfg *openAPITranscriptionConfig) {
		cfg.DocumentTitle = strings.TrimSpace(title)
	}
}

// WithOpenAPIDescription sets the generated document description.
func WithOpenAPIDescription(description string) OpenAPITranscriptionOption {
	return func(cfg *openAPITranscriptionConfig) {
		cfg.DocumentDescription = strings.TrimSpace(description)
	}
}

// WithOpenAPIVersion sets the generated document info.version value.
func WithOpenAPIVersion(version string) OpenAPITranscriptionOption {
	return func(cfg *openAPITranscriptionConfig) {
		cfg.DocumentVersion = strings.TrimSpace(version)
	}
}

// WithOpenAPIServerURL sets the public OpenAPI server URL.
func WithOpenAPIServerURL(serverURL string) OpenAPITranscriptionOption {
	return func(cfg *openAPITranscriptionConfig) {
		cfg.PublicServerURL = strings.TrimSpace(serverURL)
	}
}

// OpenAPIDocument is the document-level schema produced from endpoints.
type OpenAPIDocument struct {
	OpenAPI  string          `json:"openapi,omitempty" yaml:"openapi,omitempty"`
	Swagger  string          `json:"swagger,omitempty" yaml:"swagger,omitempty"`
	Info     OpenAPIInfo     `json:"info" yaml:"info"`
	Servers  []OpenAPIServer `json:"servers,omitempty" yaml:"servers,omitempty"`
	Host     string          `json:"host,omitempty" yaml:"host,omitempty"`
	Schemes  []string        `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Produces []string        `json:"produces,omitempty" yaml:"produces,omitempty"`
	Paths    OpenAPIPaths    `json:"paths" yaml:"paths"`
}

// OpenAPIInfo is the generated document info block.
type OpenAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version" yaml:"version"`
}

// OpenAPIServer is an OpenAPI 3 server entry.
type OpenAPIServer struct {
	URL string `json:"url" yaml:"url"`
}

// OpenAPIPaths maps OpenAPI path strings to method entries.
type OpenAPIPaths map[string]OpenAPIPathItem

// OpenAPIPathItem is the per-path OpenAPI method container.
type OpenAPIPathItem struct {
	Get  *OpenAPIOperation `json:"get,omitempty" yaml:"get,omitempty"`
	Post *OpenAPIOperation `json:"post,omitempty" yaml:"post,omitempty"`
}

// OpenAPIOperation is the minimal operation shape used by endpoint transcription.
type OpenAPIOperation struct {
	OperationID           string                            `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Summary               string                            `json:"summary,omitempty" yaml:"summary,omitempty"`
	Consumes              []string                          `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces              []string                          `json:"produces,omitempty" yaml:"produces,omitempty"`
	XGoogleBackend        *GCPGatewayBackend                `json:"x-google-backend,omitempty" yaml:"x-google-backend,omitempty"`
	XHTTPAPIInternal      bool                              `json:"x-httpapi-internal,omitempty" yaml:"x-httpapi-internal,omitempty"`
	XHTTPAPIAuthorization *authpkg.AuthorizationRequirement `json:"x-httpapi-authorization,omitempty" yaml:"x-httpapi-authorization,omitempty"`
	XHTTPAPIPriority      endpointpkg.Priority              `json:"x-httpapi-priority,omitempty" yaml:"x-httpapi-priority,omitempty"`
	Responses             map[string]OpenAPIResponse        `json:"responses" yaml:"responses"`
}

// GCPGatewayBackend is the GCP API Gateway x-google-backend extension.
type GCPGatewayBackend struct {
	Address         string   `json:"address,omitempty" yaml:"address,omitempty"`
	PathTranslation string   `json:"path_translation,omitempty" yaml:"path_translation,omitempty"`
	Deadline        *float64 `json:"deadline,omitempty" yaml:"deadline,omitempty"`
}

// OpenAPIResponse is the minimal OpenAPI response object.
type OpenAPIResponse struct {
	Description string `json:"description" yaml:"description"`
}

type EndpointTarget struct {
	endpoint httpapi.Endpoint
}

type GroupTarget struct {
	group httpapi.EndpointGroup
}

func ForEndpoint(endpoint httpapi.Endpoint) EndpointTarget {
	return EndpointTarget{endpoint: endpoint}
}

func ForGroup(group httpapi.EndpointGroup) GroupTarget {
	return GroupTarget{group: group}
}

// TranscribePublicOpenAPI emits public OpenAPI paths for the endpoint.
func (t EndpointTarget) TranscribePublicOpenAPI(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIPaths, error) {
	return t.Transcribe(OpenAPITranscriptionModePublic, opts...)
}

// TranscribeGCPGateway emits GCP API Gateway paths for the endpoint.
func (t EndpointTarget) TranscribeGCPGateway(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIPaths, error) {
	return t.Transcribe(OpenAPITranscriptionModeGCPGateway, opts...)
}

// TranscribeGateway emits the default gateway paths for the endpoint.
//
// Deprecated: use TranscribeGCPGateway when generating GCP API Gateway specs.
func (t EndpointTarget) TranscribeGateway(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIPaths, error) {
	return t.Transcribe(OpenAPITranscriptionModeGateway, opts...)
}

// TranscribePublicOpenAPIDocument emits a public OpenAPI 3.1 document.
func (t EndpointTarget) TranscribePublicOpenAPIDocument(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIDocument, error) {
	return t.TranscribeDocument(OpenAPITranscriptionModePublic, opts...)
}

// TranscribeGCPGatewayDocument emits a GCP API Gateway Swagger 2.0 document.
func (t EndpointTarget) TranscribeGCPGatewayDocument(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIDocument, error) {
	return t.TranscribeDocument(OpenAPITranscriptionModeGCPGateway, opts...)
}

// TranscribeGatewayDocument emits the default gateway document.
//
// Deprecated: use TranscribeGCPGatewayDocument when generating GCP API Gateway
// specs.
func (t EndpointTarget) TranscribeGatewayDocument(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIDocument, error) {
	return t.TranscribeDocument(OpenAPITranscriptionModeGateway, opts...)
}

// Transcribe emits OpenAPI path entries for the endpoint.
func (t EndpointTarget) Transcribe(
	mode OpenAPITranscriptionMode,
	opts ...OpenAPITranscriptionOption,
) (OpenAPIPaths, error) {
	cfg := openAPIConfig(opts)
	return transcribeEndpoint(t.endpoint, mode, cfg)
}

// TranscribeDocument emits a document-level OpenAPI schema for the endpoint.
func (t EndpointTarget) TranscribeDocument(
	mode OpenAPITranscriptionMode,
	opts ...OpenAPITranscriptionOption,
) (OpenAPIDocument, error) {
	cfg := openAPIConfig(opts)
	paths, err := transcribeEndpoint(t.endpoint, mode, cfg)
	if err != nil {
		return OpenAPIDocument{}, err
	}

	return openAPIDocument(mode, cfg, paths)
}

// TranscribePublicOpenAPI emits public OpenAPI paths for the endpoint group.
func (t GroupTarget) TranscribePublicOpenAPI(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIPaths, error) {
	return t.Transcribe(OpenAPITranscriptionModePublic, opts...)
}

// TranscribeGCPGateway emits GCP API Gateway paths for the endpoint group.
func (t GroupTarget) TranscribeGCPGateway(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIPaths, error) {
	return t.Transcribe(OpenAPITranscriptionModeGCPGateway, opts...)
}

// TranscribeGateway emits the default gateway paths for the endpoint group.
//
// Deprecated: use TranscribeGCPGateway when generating GCP API Gateway specs.
func (t GroupTarget) TranscribeGateway(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIPaths, error) {
	return t.Transcribe(OpenAPITranscriptionModeGateway, opts...)
}

// TranscribePublicOpenAPIDocument emits a public OpenAPI 3.1 document.
func (t GroupTarget) TranscribePublicOpenAPIDocument(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIDocument, error) {
	return t.TranscribeDocument(OpenAPITranscriptionModePublic, opts...)
}

// TranscribeGCPGatewayDocument emits a GCP API Gateway Swagger 2.0 document.
func (t GroupTarget) TranscribeGCPGatewayDocument(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIDocument, error) {
	return t.TranscribeDocument(OpenAPITranscriptionModeGCPGateway, opts...)
}

// TranscribeGatewayDocument emits the default gateway document.
//
// Deprecated: use TranscribeGCPGatewayDocument when generating GCP API Gateway
// specs.
func (t GroupTarget) TranscribeGatewayDocument(
	opts ...OpenAPITranscriptionOption,
) (OpenAPIDocument, error) {
	return t.TranscribeDocument(OpenAPITranscriptionModeGateway, opts...)
}

// Transcribe emits OpenAPI path entries for every endpoint in the group.
func (t GroupTarget) Transcribe(
	mode OpenAPITranscriptionMode,
	opts ...OpenAPITranscriptionOption,
) (OpenAPIPaths, error) {
	cfg := openAPIConfig(opts)
	return transcribeGroup(t.group, mode, cfg)
}

// TranscribeDocument emits a document-level OpenAPI schema for the endpoint group.
func (t GroupTarget) TranscribeDocument(
	mode OpenAPITranscriptionMode,
	opts ...OpenAPITranscriptionOption,
) (OpenAPIDocument, error) {
	cfg := openAPIConfig(opts)
	paths, err := transcribeGroup(t.group, mode, cfg)
	if err != nil {
		return OpenAPIDocument{}, err
	}

	return openAPIDocument(mode, cfg, paths)
}

func transcribeGroup(
	eg httpapi.EndpointGroup,
	mode OpenAPITranscriptionMode,
	cfg openAPITranscriptionConfig,
) (OpenAPIPaths, error) {
	prefix, err := joinOpenAPIPath(cfg.PathPrefix, eg.PathPrefix)
	if err != nil {
		return nil, err
	}
	cfg.PathPrefix = prefix

	paths := OpenAPIPaths{}
	for _, endpoint := range eg.ResolvedEndpoints() {
		if mode == OpenAPITranscriptionModePublic && endpoint.IsInternal() {
			continue
		}

		endpointPaths, err := transcribeEndpoint(endpoint, mode, cfg)
		if err != nil {
			return nil, err
		}
		if err := mergeOpenAPIPaths(paths, endpointPaths); err != nil {
			return nil, err
		}
	}

	return paths, nil
}

func transcribeEndpoint(
	e httpapi.Endpoint,
	mode OpenAPITranscriptionMode,
	cfg openAPITranscriptionConfig,
) (OpenAPIPaths, error) {
	if mode == OpenAPITranscriptionModePublic && e.IsInternal() {
		return nil, ErrInternalEndpointPublicOpenAPI
	}

	path, err := joinOpenAPIPath(cfg.PathPrefix, e.Pattern())
	if err != nil {
		return nil, err
	}

	operation, err := openAPIOperation(e, mode, path, cfg)
	if err != nil {
		return nil, err
	}

	item := OpenAPIPathItem{}
	if err := item.setOperation(e.Method(), operation); err != nil {
		return nil, err
	}

	return OpenAPIPaths{path: item}, nil
}

func openAPIOperation(
	e httpapi.Endpoint,
	mode OpenAPITranscriptionMode,
	path string,
	cfg openAPITranscriptionConfig,
) (OpenAPIOperation, error) {
	route := e.RouteSpec()
	operation := OpenAPIOperation{
		OperationID: defaultOperationID(e.Method(), path),
		Summary:     fmt.Sprintf("%s %s", e.Method(), path),
		Responses:   placeholderResponses(),
	}
	if route.OperationID != "" {
		operation.OperationID = route.OperationID
	}
	if route.Summary != "" {
		operation.Summary = route.Summary
	}

	if e.IsInternal() {
		operation.XHTTPAPIInternal = true
	}
	if auth := e.Authorization(); auth.Required {
		operation.XHTTPAPIAuthorization = &auth
	}
	if priority := e.Priority(); priority != "" {
		operation.XHTTPAPIPriority = priority
	}

	switch mode {
	case OpenAPITranscriptionModePublic:
		return operation, nil
	case OpenAPITranscriptionModeGCPGateway:
		backend, err := gcpGatewayBackend(
			routeBackendWithDefaults(route.Backend, RouteBackend{
				Address:  cfg.GatewayBackendAddress,
				PathMode: RoutePathModeAppend,
			}),
		)
		if err != nil {
			return OpenAPIOperation{}, err
		}
		operation.Consumes = contentTypesForOpenAPI(e.AcceptedContentTypes())
		operation.Produces = []string{ApplicationJson}
		operation.XGoogleBackend = &backend
		return operation, nil
	default:
		return OpenAPIOperation{}, fmt.Errorf(
			"httpapi: unsupported openapi transcription mode %q",
			mode,
		)
	}
}

func contentTypesForOpenAPI(contentTypes []ContentType) []string {
	contentTypes = normalizeContentTypes(contentTypes)
	values := make([]string, 0, len(contentTypes))
	for _, contentType := range contentTypes {
		values = append(values, string(contentType))
	}
	return values
}

func normalizeContentTypes(contentTypes []ContentType) []ContentType {
	return endpointpkg.NormalizeContentTypeSlice(contentTypes)
}

func routeBackendWithDefaults(backend, defaults RouteBackend) RouteBackend {
	return backend.WithDefaults(defaults)
}

func gcpGatewayBackend(backend RouteBackend) (GCPGatewayBackend, error) {
	backend = endpointpkg.NormalizeRouteBackend(backend)
	if backend.Address == "" {
		return GCPGatewayBackend{}, ErrGCPGatewayBackendAddressRequired
	}

	pathTranslation := gcpGatewayPathTranslationAppend
	switch backend.PathMode {
	case "", RoutePathModeAppend:
		pathTranslation = gcpGatewayPathTranslationAppend
	case RoutePathModeConstant:
		pathTranslation = gcpGatewayPathTranslationConstant
	default:
		return GCPGatewayBackend{}, fmt.Errorf(
			"httpapi: unsupported route path mode %q",
			backend.PathMode,
		)
	}

	deadline, err := gcpGatewayBackendDeadline(backend)
	if err != nil {
		return GCPGatewayBackend{}, err
	}

	return GCPGatewayBackend{
		Address:         backend.Address,
		PathTranslation: pathTranslation,
		Deadline:        deadline,
	}, nil
}

func gcpGatewayBackendDeadline(backend RouteBackend) (*float64, error) {
	if backend.Timeout == 0 {
		return nil, nil
	}
	if backend.Timeout > gcpGatewayBackendDeadlineMax {
		return nil, fmt.Errorf(
			"%w: timeout=%s max=%s",
			ErrGCPGatewayBackendDeadlineExceeded,
			backend.Timeout,
			gcpGatewayBackendDeadlineMax,
		)
	}

	seconds := backend.Timeout.Seconds()
	return &seconds, nil
}

func openAPIConfig(opts []OpenAPITranscriptionOption) openAPITranscriptionConfig {
	cfg := openAPITranscriptionConfig{
		DocumentTitle:       defaultOpenAPIDocumentTitle,
		DocumentDescription: defaultOpenAPIDocumentDescription,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

func openAPIDocument(
	mode OpenAPITranscriptionMode,
	cfg openAPITranscriptionConfig,
	paths OpenAPIPaths,
) (OpenAPIDocument, error) {
	if strings.TrimSpace(cfg.DocumentVersion) == "" {
		return OpenAPIDocument{}, ErrOpenAPIDocumentVersionRequired
	}

	doc := OpenAPIDocument{
		Info: OpenAPIInfo{
			Title:       cfg.DocumentTitle,
			Description: cfg.DocumentDescription,
			Version:     cfg.DocumentVersion,
		},
		Paths: paths,
	}

	switch mode {
	case OpenAPITranscriptionModePublic:
		doc.OpenAPI = defaultOpenAPIDocumentSpecVersion
		if cfg.PublicServerURL != "" {
			doc.Servers = []OpenAPIServer{{URL: cfg.PublicServerURL}}
		}
		return doc, nil
	case OpenAPITranscriptionModeGCPGateway:
		doc.Swagger = defaultGCPGatewayDocumentSpecVersion
		doc.Host = cfg.GatewayHost
		doc.Schemes = []string{defaultGCPGatewayScheme}
		doc.Produces = []string{ApplicationJson}
		return doc, nil
	default:
		return OpenAPIDocument{}, fmt.Errorf(
			"httpapi: unsupported openapi transcription mode %q",
			mode,
		)
	}
}

func mergeOpenAPIPaths(into, paths OpenAPIPaths) error {
	for path, item := range paths {
		existing := into[path]
		if err := existing.merge(item); err != nil {
			return fmt.Errorf("httpapi: merge openapi path %q: %w", path, err)
		}
		into[path] = existing
	}

	return nil
}

func (p *OpenAPIPathItem) setOperation(method HttpMethod, operation OpenAPIOperation) error {
	switch method {
	case GET:
		if p.Get != nil {
			return fmt.Errorf("duplicate %s operation", method)
		}
		p.Get = &operation
	case POST:
		if p.Post != nil {
			return fmt.Errorf("duplicate %s operation", method)
		}
		p.Post = &operation
	default:
		return fmt.Errorf("unsupported openapi method %q", method)
	}

	return nil
}

func (p *OpenAPIPathItem) merge(other OpenAPIPathItem) error {
	if other.Get != nil {
		if p.Get != nil {
			return fmt.Errorf("duplicate %s operation", GET)
		}
		p.Get = other.Get
	}
	if other.Post != nil {
		if p.Post != nil {
			return fmt.Errorf("duplicate %s operation", POST)
		}
		p.Post = other.Post
	}

	return nil
}

func joinOpenAPIPath(prefix, pattern string) (string, error) {
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
		return "", fmt.Errorf("httpapi: join openapi path: %w", err)
	}
	path, err = url.PathUnescape(path)
	if err != nil {
		return "", fmt.Errorf("httpapi: unescape openapi path: %w", err)
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

func placeholderResponses() map[string]OpenAPIResponse {
	return map[string]OpenAPIResponse{
		"default": {Description: placeholderResponseDescription},
	}
}

func defaultOperationID(method HttpMethod, path string) string {
	parts := []string{strings.ToLower(method)}
	for _, part := range strings.FieldsFunc(path, openAPIOperationIDSeparator) {
		part = strings.Trim(part, "{}")
		if part == "" {
			continue
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, "_")
}

func openAPIOperationIDSeparator(r rune) bool {
	return r == '/' || r == '-' || r == '_' || r == '.' || r == ':'
}
