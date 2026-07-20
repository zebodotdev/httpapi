package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	contentTypeHeaderKey = "content-type"
	xReqIDHeaderKey      = "x-request-id"
	xReqTimingHeaderKey  = "x-request-timing"

	corsOriginHeaderKey  = "access-control-allow-origin"
	corsMethodsHeaderKey = "access-control-allow-methods"
	corsHeadersHeaderKey = "access-control-allow-headers"
)

var logr = log.New(os.Stdout, "[httpapi/response]: ", log.Flags()|log.Llongfile)

// WriteOptions controls how a Res is written to the HTTP connection.
type WriteOptions struct {
	// RequestID is written as x-request-id.
	RequestID string

	// Duration is written as x-request-timing.
	Duration time.Duration

	// WriteTimeout bounds the response write. Zero disables the write deadline.
	WriteTimeout time.Duration
}

// WriteResult describes the completed response write.
type WriteResult struct {
	// Status is the response status code written to the client.
	Status int

	// BytesWritten is the number of bytes written or, for some writers, the
	// encoded body length when the writer reports zero bytes.
	BytesWritten int

	// Streamed reports whether the response was copied from BodyReader.
	Streamed bool
}

// WriteResponse writes a response using httpapi's standard headers, CORS
// defaults, body encoding, streaming support, and write deadline handling.
func WriteResponse(
	w http.ResponseWriter,
	res *Res,
	opts WriteOptions,
) (WriteResult, error) {
	if res == nil {
		return WriteResult{}, fmt.Errorf("httpapi: response is required")
	}

	if res.BodyReader != nil {
		return WriteResponseStream(w, res, opts)
	}

	body, err := EncodeResponseBody(res)
	if err != nil {
		return WriteResult{}, err
	}

	return WriteResponseBody(w, res, body, opts)
}

// EncodeResponseBody encodes a non-streaming response body.
//
// JSON responses are encoded with encoding/json. TextHTML and TextPlain
// responses use the string body directly when possible.
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
//
// It applies standard headers, optional write deadlines, and returns the write
// result for audit and logging.
func WriteResponseBody(
	w http.ResponseWriter,
	res *Res,
	body []byte,
	opts WriteOptions,
) (WriteResult, error) {
	writeDeadlineSet := setEndpointWriteDeadline(w, opts.RequestID, opts.WriteTimeout)
	if writeDeadlineSet {
		defer clearEndpointWriteDeadline(w, opts.RequestID)
	}

	writeResponseHeader(w, res, opts)
	written, err := w.Write(body)
	if written == 0 {
		written = len(body)
	}

	return WriteResult{
		Status:       res.Status,
		BytesWritten: written,
	}, err
}

// WriteResponseStream writes a streaming response body.
//
// If BodyReader implements io.Closer, it is closed after streaming completes.
func WriteResponseStream(
	w http.ResponseWriter,
	res *Res,
	opts WriteOptions,
) (WriteResult, error) {
	writeDeadlineSet := setEndpointWriteDeadline(w, opts.RequestID, opts.WriteTimeout)
	if writeDeadlineSet {
		defer clearEndpointWriteDeadline(w, opts.RequestID)
	}

	writeResponseHeader(w, res, opts)
	if closer, ok := res.BodyReader.(io.Closer); ok {
		defer closer.Close()
	}
	written, err := io.Copy(w, res.BodyReader)
	return WriteResult{
		Status:       res.Status,
		BytesWritten: int(written),
		Streamed:     true,
	}, err
}

func writeResponseHeader(
	w http.ResponseWriter,
	res *Res,
	opts WriteOptions,
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

type endpointDeadlineSetter func(*http.ResponseController, time.Time) error

func setEndpointWriteDeadline(
	w http.ResponseWriter,
	requestID string,
	timeout time.Duration,
) bool {
	if timeout < 0 {
		panic("httpapi/response: write timeout cannot be negative")
	}
	if timeout == 0 {
		return false
	}

	return applyEndpointDeadline(
		w,
		requestID,
		time.Now().Add(timeout),
		"write",
		(*http.ResponseController).SetWriteDeadline,
	)
}

func clearEndpointWriteDeadline(w http.ResponseWriter, requestID string) {
	applyEndpointDeadline(
		w,
		requestID,
		time.Time{},
		"write",
		(*http.ResponseController).SetWriteDeadline,
	)
}

func applyEndpointDeadline(
	w http.ResponseWriter,
	requestID string,
	deadline time.Time,
	phase string,
	set endpointDeadlineSetter,
) bool {
	if w == nil {
		return false
	}

	if err := set(http.NewResponseController(w), deadline); err != nil {
		if errors.Is(err, http.ErrNotSupported) {
			return false
		}

		logr.Printf(
			"failed to set endpoint %s deadline:"+
				" request_id=%s deadline=%v error=%v",
			phase,
			requestID,
			deadline,
			err,
		)
		return false
	}

	return true
}
