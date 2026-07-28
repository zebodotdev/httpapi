package response

import "net/http"

func cloneHeader(header http.Header) http.Header {
	if header == nil {
		return nil
	}
	return header.Clone()
}

func mergeHeader(dst http.Header, src http.Header) http.Header {
	if len(src) == 0 {
		return cloneHeader(dst)
	}
	merged := cloneHeader(dst)
	if merged == nil {
		merged = http.Header{}
	}
	for key, values := range src {
		for _, value := range values {
			merged.Add(key, value)
		}
	}
	return merged
}
