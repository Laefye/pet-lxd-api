package lxd

import (
	"context"
	"errors"
	"net/http"
)

type InstanceSource struct {
	Alias string `json:"alias"`
	Type  string `json:"type"`
}

type Resources struct {
	Instances []string `json:"instances"`
}

type ExecSubMetadata struct {
	Fds map[string]string `json:"fds"`
}

type BaseMetadata struct {
	Class     string    `json:"class"`
	Id        string    `json:"id"`
	Resources Resources `json:"resources"`
}

type ExecMetadata struct {
	BaseMetadata
	Metadata ExecSubMetadata `json:"metadata"`
}

type AddressInfo struct {
	Family  string `json:"family"`
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Scope   string `json:"scope"`
}

type NetworkInfo struct {
	Addresses []AddressInfo `json:"addresses"`
}

type StateMetadata struct {
	BaseMetadata
	Processes int                    `json:"processes"`
	Network   map[string]NetworkInfo `json:"network"`
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

type Fd struct {
	Stdin  *WebSocketStream
	Stdout *WebSocketStream
	Stderr *WebSocketStream
}

var ErrMissingWebSocketFlag = errors.New("wait-for-websocket must be true")

func (r *Rest) Exec(ctx context.Context, path Path, req ExecRequest) (*Fd, error) {
	if req.WaitForWS == false {
		return nil, ErrMissingWebSocketFlag
	}
	res, execMetadata, err := R[ExecMetadata](r, ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	operationPath, err := ParsePath(res.Operation)
	if err != nil {
		return nil, err
	}
	wsPath := operationPath.Join("websocket")
	stdin, err := r.WebSocket(ctx, wsPath.WithSecret(execMetadata.Metadata.Fds["0"]))
	if err != nil {
		return nil, err
	}
	stdout, err := r.WebSocket(ctx, wsPath.WithSecret(execMetadata.Metadata.Fds["1"]))
	if err != nil {
		stdin.Close()
		return nil, err
	}
	stderr, err := r.WebSocket(ctx, wsPath.WithSecret(execMetadata.Metadata.Fds["2"]))
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, err
	}
	return &Fd{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}, nil
}

func (f *Fd) Close() {
	f.Stdout.Close()
	f.Stderr.Close()
	f.Stdin.Close()
}
