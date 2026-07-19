package endpoint

import (
	"fmt"
	"maps"
	"strings"
)

type Method = string
type ContentType = string

const (
	OPTIONS Method = "OPTIONS"
	POST    Method = "POST"
	GET     Method = "GET"

	ApplicationJson           ContentType = "application/json"
	ApplicationFormURLEncoded ContentType = "application/x-www-form-urlencoded"
	MultipartFormData         ContentType = "multipart/form-data"
	TextHTML                  ContentType = "text/html"
	TextPlain                 ContentType = "text/plain; charset=utf-8"
)

func NormalizeMethod(method Method) Method {
	normalized := Method(strings.ToUpper(strings.TrimSpace(string(method))))
	switch normalized {
	case POST, GET:
		return normalized
	default:
		panic(
			"invalid method for this http endpoint." +
				" supported methods are `GET` and `POST`",
		)
	}
}

func NormalizeContentType(contentType ContentType) ContentType {
	contentType = ContentType(strings.TrimSpace(string(contentType)))
	if contentType == "" {
		return ApplicationJson
	}

	return contentType
}

func NormalizeContentTypes(primary ContentType, additional ...ContentType) []ContentType {
	var contentTypes []ContentType
	seen := map[ContentType]bool{}
	if strings.TrimSpace(string(primary)) != "" || len(additional) == 0 {
		contentType := NormalizeContentType(primary)
		contentTypes = append(contentTypes, contentType)
		seen[contentType] = true
	}
	for _, contentType := range additional {
		contentType = NormalizeContentType(contentType)
		if seen[contentType] {
			continue
		}
		seen[contentType] = true
		contentTypes = append(contentTypes, contentType)
	}

	return contentTypes
}

func NormalizeContentTypeSlice(contentTypes []ContentType) []ContentType {
	if len(contentTypes) == 0 {
		return []ContentType{ApplicationJson}
	}

	normalized := make([]ContentType, 0, len(contentTypes))
	seen := map[ContentType]bool{}
	for _, contentType := range contentTypes {
		contentType = NormalizeContentType(contentType)
		if seen[contentType] {
			continue
		}
		seen[contentType] = true
		normalized = append(normalized, contentType)
	}
	return normalized
}

func PrimaryContentType(contentTypes []ContentType) ContentType {
	if len(contentTypes) == 0 {
		return ApplicationJson
	}
	return NormalizeContentType(contentTypes[0])
}

func CloneContentTypes(contentTypes []ContentType) []ContentType {
	if len(contentTypes) == 0 {
		return nil
	}
	return append([]ContentType(nil), contentTypes...)
}

func JoinContentTypes(contentTypes []ContentType) string {
	if len(contentTypes) == 0 {
		return string(ApplicationJson)
	}

	parts := make([]string, 0, len(contentTypes))
	for _, contentType := range contentTypes {
		parts = append(parts, string(NormalizeContentType(contentType)))
	}
	return strings.Join(parts, ", ")
}

func CloneAuthKeys(keys map[string]bool) map[string]bool {
	if len(keys) == 0 {
		return nil
	}

	cloned := make(map[string]bool, len(keys))
	maps.Copy(cloned, keys)

	return cloned
}

func ValidateContentType(actual ContentType, expected []ContentType) error {
	actual = NormalizeContentType(actual)
	expected = NormalizeContentTypeSlice(expected)
	for _, candidate := range expected {
		if strings.Contains(string(actual), string(candidate)) {
			return nil
		}
	}

	return fmt.Errorf("httpapi: content type %q does not match any of %q", actual, JoinContentTypes(expected))
}
