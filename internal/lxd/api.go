package lxd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

type Rest struct {
	Client *http.Client
	Dialer *websocket.Dialer
	base   Path
	Host   string
}

var defaultBase = Path{
	Segments: []string{"1.0"},
}

func newRest(host string, client *http.Client, dialer *websocket.Dialer, base Path) *Rest {
	return &Rest{
		Client: client,
		Dialer: dialer,
		base:   base,
		Host:   host,
	}
}

func NewRest(host string, client *http.Client, dialer *websocket.Dialer) *Rest {
	return newRest(host, client, dialer, defaultBase)
}

func NewRestWithProject(host string, client *http.Client, dialer *websocket.Dialer, project string) *Rest {
	return newRest(host, client, dialer, defaultBase.withQuery("project", project))
}

type response struct {
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

func (r *Rest) httpsPath(path Path) string {
	return "https://" + string(r.Host) + path.String()
}

func (r *Rest) do(ctx context.Context, method string, path Path, data io.Reader, header http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, r.httpsPath(path), data)
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

func parseResponse(reader io.Reader) (*response, error) {
	var out response
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

func parseMetadata[T any](r *response) (*T, error) {
	var out T
	if err := json.Unmarshal(r.Metadata, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *Rest) request(ctx context.Context, method string, path Path, data interface{}) (*response, error) {
	req, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	res, err := r.upload(ctx, method, path, bytes.NewReader(req), http.Header{"Content-Type": []string{"application/json"}})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func request[T any](r *Rest, ctx context.Context, method string, path Path, data interface{}) (*response, *T, error) {
	resp, err := r.request(ctx, method, path, data)
	if err != nil {
		return nil, nil, err
	}
	out, err := parseMetadata[T](resp)
	if err != nil {
		return resp, nil, err
	}
	return resp, out, nil
}

func (r *Rest) upload(ctx context.Context, method string, path Path, reader io.Reader, header http.Header) (*response, error) {
	resp, err := r.do(ctx, method, path, reader, header)
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
