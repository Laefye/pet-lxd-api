package lxd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Rest struct {
	Client   http.Client
	Endpoint Endpoint
}

type BaseResponse struct {
	Type       string `json:"type"`
	ErrorCode  int    `json:"error_code"`
	Error      string `json:"error"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
}

type Resources struct {
	Instances []string `json:"instances"`
}

type BaseMetadata struct {
	Class     string    `json:"class"`
	Id        string    `json:"id"`
	Resources Resources `json:"resources"`
}

func (r *Rest) Get(ctx context.Context, path string, out interface{}) error {
	fmt.Println("GET", r.Endpoint.Https(path))
	req, err := http.NewRequestWithContext(ctx, "GET", r.Endpoint.Https(path), nil)
	if err != nil {
		return err
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (r *Rest) Post(ctx context.Context, path string, data interface{}, out interface{}) error {
	bodyJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.Endpoint.Https(path), bytes.NewReader(bodyJson))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
