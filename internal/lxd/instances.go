package lxd

import (
	"context"
	"net/http"
)

const (
	InstanceTypeContainer = "container"
	InstanceTypeVM        = "virtual-machine"
	SourceTypeImage       = "image"
)

type InstanceSource struct {
	Alias string `json:"alias"`
	Type  string `json:"type"`
}

type resources struct {
	Instances []string `json:"instances"`
}

type execSubMetadata struct {
	Fds map[string]string `json:"fds"`
}

type baseMetadata struct {
	Class string `json:"class"`
	Id    string `json:"id"`
}

type execMetadata struct {
	baseMetadata
	Metadata execSubMetadata `json:"metadata"`
}

type resourcedMetadata struct {
	baseMetadata
	Resources resources `json:"resources"`
}

type addressInfo struct {
	Family  string `json:"family"`
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Scope   string `json:"scope"`
}

type networkInfo struct {
	Addresses []addressInfo `json:"addresses"`
}

type stateMetadata struct {
	baseMetadata
	Status    string                 `json:"status"`
	Processes int                    `json:"processes"`
	Network   map[string]networkInfo `json:"network"`
}

type device struct {
	Type   string `json:"type"`
	Size   string `json:"size,omitempty"`
	Source string `json:"source,omitempty"`
	Path   string `json:"path,omitempty"`
	Pool   string `json:"pool,omitempty"`
	IoBus  string `json:"io.bus,omitempty"`
}

type InstanceCreationRequest struct {
	Name    string            `json:"name,omitempty"`
	Source  InstanceSource    `json:"source"`
	Start   bool              `json:"start"`
	Type    string            `json:"type,omitempty"`
	Config  map[string]string `json:"config,omitempty"`
	Devices map[string]device `json:"devices,omitempty"`
}

type ExecRequest struct {
	Command      []string          `json:"command"`
	WaitForWS    bool              `json:"wait-for-websocket,omitempty"`
	Interactive  bool              `json:"interactive,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	RecordOutput bool              `json:"record-output,omitempty"`
	Cwd          string            `json:"cwd,omitempty"`
	Group        int               `json:"group,omitempty"`
	User         int               `json:"user,omitempty"`
}

type InstanceCreateArgs struct {
	Name   string
	Source InstanceSource
	Start  bool
	Type   string
}

type Instance struct {
	rest *Rest
	path Path
}

func (r *Rest) CreateInstance(ctx context.Context, req InstanceCreationRequest) (*Instance, error) {
	res, _, err := request[resourcedMetadata](r, ctx, http.MethodPost, r.base.Join("instances"), req)
	if err != nil {
		return nil, err
	}
	operationPath, err := ParsePath(res.Operation)
	if err != nil {
		return nil, err
	}
	res, metadata, err := request[resourcedMetadata](r, ctx, http.MethodGet, operationPath.Join("wait"), nil)
	if err != nil {
		return nil, err
	}
	instancePath, err := ParsePath(metadata.Resources.Instances[0])
	if err != nil {
		return nil, err
	}
	return &Instance{rest: r, path: *instancePath}, nil
}

func (r *Rest) Instance(ctx context.Context, name string) (*Instance, error) {
	path := r.base.Join("instances").Join(name)
	return &Instance{rest: r, path: path}, nil
}

type State struct {
	Status    string
	Processes int
}

func (i *Instance) GetState(ctx context.Context) (*State, error) {
	_, metadata, err := request[stateMetadata](i.rest, ctx, http.MethodGet, i.path.Join("state"), nil)
	if err != nil {
		return nil, err
	}
	return &State{
		Status:    metadata.Status,
		Processes: metadata.Processes,
	}, nil
}

func (i *Instance) Exec(ctx context.Context, req ExecRequest) (WebSockets, error) {
	res, metadata, err := request[execMetadata](i.rest, ctx, http.MethodPost, i.path.Join("exec"), req)
	if err != nil {
		return nil, err
	}
	websocketPath, err := ParsePath(res.Operation)
	if err != nil {
		return nil, err
	}
	streams := make(WebSockets)
	for name, path := range metadata.Metadata.Fds {
		stream, err := i.rest.webSocket(ctx, websocketPath.Join("websocket").withQuery("secret", path))
		if err != nil {
			return nil, err
		}
		streams[name] = stream
	}
	return streams, nil
}
