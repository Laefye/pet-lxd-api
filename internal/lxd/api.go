package lxd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Rest struct {
	Client   *http.Client
	Endpoint Endpoint
}

type Response struct {
	Type       string          `json:"type"`
	ErrorCode  int             `json:"error_code"`
	Error      string          `json:"error"`
	Status     string          `json:"status"`
	StatusCode int             `json:"status_code"`
	Metadata   json.RawMessage `json:"metadata"`
	Operation  string          `json:"operation"`
}

type LxdApiError struct {
	Code    int
	Message string
}

func (e *LxdApiError) Error() string {
	return e.Message
}

func (r *Rest) Do(ctx context.Context, method, path string, data io.Reader, header http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, r.Endpoint.Https(path), data)
	if err != nil {
		return nil, err
	}
	for key, values := range header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func parseResponse(reader io.Reader) (*Response, error) {
	var out Response
	if err := json.NewDecoder(reader).Decode(&out); err != nil {
		return nil, err
	}
	if out.ErrorCode != 0 {
		return &out, &LxdApiError{
			Code:    out.ErrorCode,
			Message: out.Error,
		}
	}
	return &out, nil
}

func (r *Rest) Request(ctx context.Context, method, path string, data interface{}) (*Response, error) {
	req, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	resp, err := r.Do(ctx, method, path, bytes.NewReader(req), http.Header{
		"Content-Type": []string{"application/json"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out, err := parseResponse(resp.Body)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func R[T any](r *Rest, ctx context.Context, method, path string, data interface{}) (*Response, *T, error) {
	resp, err := r.Request(ctx, method, path, data)
	if err != nil {
		return nil, nil, err
	}
	var out T
	if err := json.Unmarshal(resp.Metadata, &out); err != nil {
		return nil, nil, err
	}
	return resp, &out, nil
}

func (r *Rest) Upload(ctx context.Context, path string, reader io.Reader, header http.Header) (*Response, error) {
	resp, err := r.Do(ctx, http.MethodPost, path, reader, header)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out, err := parseResponse(resp.Body)
	if err != nil {
		return nil, err
	}
	return out, nil
}
