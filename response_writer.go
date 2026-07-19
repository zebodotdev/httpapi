package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ResponseWriteOptions controls how a Res is written to the HTTP connection.
type ResponseWriteOptions struct {
	RequestID string
	Duration  time.Duration
	Timeout   EndpointTimeoutSpec
}

// ResponseWriteResult describes the completed response write.
type ResponseWriteResult struct {
	Status       int
	BytesWritten int
	Streamed     bool
}

// WriteResponse writes a response using httpapi's standard headers, CORS
// defaults, body encoding, streaming support, and write deadline handling.
func WriteResponse(
	w http.ResponseWriter,
	res *Res,
	opts ResponseWriteOptions,
) (ResponseWriteResult, error) {
	if res == nil {
		return ResponseWriteResult{}, fmt.Errorf("httpapi: response is required")
	}

	if res.BodyReader != nil {
		return WriteResponseStream(w, res, opts)
	}

	body, err := EncodeResponseBody(res)
	if err != nil {
		return ResponseWriteResult{}, err
	}

	return WriteResponseBody(w, res, body, opts)
}

// EncodeResponseBody encodes a non-streaming response body.
func EncodeResponseBody(res *Res) ([]byte, error) {
	if res == nil {
		return nil, fmt.Errorf("httpapi: response is required")
	}
	if res.BodyReader != nil {
		return nil, fmt.Errorf("httpapi: streamed response body cannot be buffered")
	}
	if res.ContentType == TextHTML || res.ContentType == TextPlain {
		if body, ok := res.Body.(string); ok {
			return []byte(body), nil
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(res.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// WriteResponseBody writes a pre-encoded non-streaming response body.
func WriteResponseBody(
	w http.ResponseWriter,
	res *Res,
	body []byte,
	opts ResponseWriteOptions,
) (ResponseWriteResult, error) {
	writeDeadlineSet := setEndpointWriteDeadline(w, opts.RequestID, opts.Timeout)
	if writeDeadlineSet {
		defer clearEndpointWriteDeadline(w, opts.RequestID)
	}

	writeResponseHeader(w, res, opts)
	written, err := w.Write(body)
	if written == 0 {
		written = len(body)
	}

	return ResponseWriteResult{
		Status:       res.Status,
		BytesWritten: written,
	}, err
}

// WriteResponseStream writes a streaming response body.
func WriteResponseStream(
	w http.ResponseWriter,
	res *Res,
	opts ResponseWriteOptions,
) (ResponseWriteResult, error) {
	writeDeadlineSet := setEndpointWriteDeadline(w, opts.RequestID, opts.Timeout)
	if writeDeadlineSet {
		defer clearEndpointWriteDeadline(w, opts.RequestID)
	}

	writeResponseHeader(w, res, opts)
	if closer, ok := res.BodyReader.(io.Closer); ok {
		defer closer.Close()
	}
	written, err := io.Copy(w, res.BodyReader)
	return ResponseWriteResult{
		Status:       res.Status,
		BytesWritten: int(written),
		Streamed:     true,
	}, err
}

func writeResponseHeader(
	w http.ResponseWriter,
	res *Res,
	opts ResponseWriteOptions,
) {
	rsh := w.Header()
	for k, v := range res.Header {
		for _, hv := range v {
			rsh.Add(k, hv)
		}
	}

	rsh.Add(contentTypeHeaderKey, res.ContentType)
	rsh.Add(xReqTimingHeaderKey, opts.Duration.String())
	rsh.Add(xReqIDHeaderKey, opts.RequestID)
	rsh.Add(corsOriginHeaderKey, "*")
	rsh.Add(corsMethodsHeaderKey, "*")
	rsh.Add(corsHeadersHeaderKey, "*")
	w.WriteHeader(res.Status)

	res.Header = rsh
}
