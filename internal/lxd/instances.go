package lxd

import (
	"context"
	"encoding/json"
	"fmt"
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

type CreateInstanceResponse struct {
	BaseResponse
	Metadata BaseMetadata `json:"metadata"`
}

func (r *Rest) CreateInstance(ctx context.Context, req CreateInstanceRequest) (*CreateInstanceResponse, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("CreateInstanceRequest: %s\n", string(jsonReq))
	var resp CreateInstanceResponse
	err = r.Post(ctx, "/1.0/instances", req, &resp)
	return &resp, err
}

type GetInstanceMetadataState struct {
	BaseMetadata
	Processes int `json:"processes"`
}

type GetInstancesStateResponse struct {
	BaseResponse
	Metadata GetInstanceMetadataState `json:"metadata"`
}

func (r *Rest) GetInstanceState(ctx context.Context, name string) (*GetInstancesStateResponse, error) {
	var instance GetInstancesStateResponse
	err := r.Get(ctx, fmt.Sprintf("/1.0/instances/%s/state", name), &instance)
	return &instance, err
}
