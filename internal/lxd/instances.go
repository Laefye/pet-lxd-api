package lxd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type InstanceSource struct {
	Alias string `json:"alias"`
	Type  string `json:"type"`
}

const (
	InstanceTypeContainer = "container"
	InstanceTypeVM        = "virtual-machine"
)

type Device struct {
	Type   string `json:"type"`
	Size   string `json:"size,omitempty"`
	Source string `json:"source,omitempty"`
	Path   string `json:"path,omitempty"`
	Pool   string `json:"pool,omitempty"`
	IoBus  string `json:"io.bus,omitempty"`
}

type CreateInstanceRequest struct {
	Name    *string           `json:"name,omitempty"`
	Source  InstanceSource    `json:"source"`
	Start   bool              `json:"start"`
	Type    string            `json:"type,omitempty"`
	Config  map[string]string `json:"config,omitempty"`
	Devices map[string]Device `json:"devices,omitempty"`
}

func (r *Rest) CreateInstance(ctx context.Context, req CreateInstanceRequest) (*RestResponse, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("CreateInstanceRequest: %s\n", string(jsonReq))
	return r.Request(ctx, http.MethodPost, "/1.0/instances", req)
}

func (r *Rest) GetInstanceState(ctx context.Context, name string) (*RestResponse, error) {
	return r.Request(ctx, http.MethodGet, fmt.Sprintf("/1.0/instances/%s/state", name), nil)
}

type ExecRequest struct {
	Command      []string          `json:"command"`
	WaitForWS    bool              `json:"wait-for-websocket"`
	Interactive  bool              `json:"interactive"`
	Environment  map[string]string `json:"environment,omitempty"`
	RecordOutput bool              `json:"record-output,omitempty"`
}

func (r *Rest) ExecInstance(ctx context.Context, name string, exec ExecRequest) (*RestResponse, error) {
	return r.Request(ctx, http.MethodPost, fmt.Sprintf("/1.0/instances/%s/exec", name), exec)
}
