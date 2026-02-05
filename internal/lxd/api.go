package lxd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type Rest struct {
	Client   http.Client
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

type Resources struct {
	Instances []string `json:"instances"`
}

type Metadata struct {
	Fds map[string]string `json:"fds"`
}

type RestMetadata struct {
	Class     string    `json:"class"`
	Id        string    `json:"id"`
	Resources Resources `json:"resources"`
	Metadata  Metadata  `json:"metadata"`
	Processes int       `json:"processes"`
}

func (r *Rest) createRequest(ctx context.Context, method, path string, data interface{}) (*http.Request, error) {
	if data != nil {
		bodyJson, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, method, r.Endpoint.Https(path), bytes.NewReader(bodyJson))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}
	return http.NewRequestWithContext(ctx, method, r.Endpoint.Https(path), nil)
}

func (r *Rest) Request(ctx context.Context, method, path string, data interface{}) (*Response, error) {
	req, err := r.createRequest(ctx, method, path, data)
	if err != nil {
		return nil, err
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out Response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
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
