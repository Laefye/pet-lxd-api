package lxd

import (
	"context"
	"net/http"
)

const (
	ProtocolTCP = "tcp"
	ProtocolUDP = "udp"
)

type ForwardPort struct {
	ListenPort string `json:"listen_port"`
	TargetPort string `json:"target_port"`
	TargetAddr string `json:"target_address"`
	Protocol   string `json:"protocol"`
}

type CreateForwardAddressRequest struct {
	ListenAddress string        `json:"listen_address"`
	Ports         []ForwardPort `json:"ports"`
}

type ForwardAddressInfo struct {
	ListenAddress string        `json:"listen_address"`
	Ports         []ForwardPort `json:"ports"`
}

type Network struct {
	rest *Rest
	path resourcePath
}

func (r *Rest) Network(name string) (*Network, error) {
	path := r.base.join("networks").withoutQuery("project").join(name)
	return &Network{rest: r, path: path}, nil
}

type ForwardAddress struct {
	rest *Rest
	path resourcePath
}

func (f *Network) ForwardAddress(listenAddress string) (*ForwardAddress, error) {
	path := f.path.join("forwards").join(listenAddress)
	return &ForwardAddress{rest: f.rest, path: path}, nil
}

func (n *Network) CreateForwardAddress(ctx context.Context, req CreateForwardAddressRequest) (*ForwardAddress, error) {
	_, err := n.rest.request(ctx, http.MethodPost, n.path.join("forwards"), req)
	if err != nil {
		return nil, err
	}
	return &ForwardAddress{rest: n.rest, path: n.path.join("forwards").join(req.ListenAddress)}, nil
}

func (n *Network) GetForwardAddresses(ctx context.Context) ([]ForwardAddress, error) {
	_, pathes, err := request[[]string](n.rest, ctx, http.MethodGet, n.path.join("forwards"), nil)
	if err != nil {
		return nil, err
	}
	var out []ForwardAddress
	for _, path := range *pathes {
		path, err := ParsePath(path)
		if err != nil {
			return nil, err
		}
		out = append(out, ForwardAddress{
			rest: n.rest,
			path: *path,
		})
	}
	return out, nil
}

type forwardAddressMetadata struct {
	baseMetadata
	ListenAddress string        `json:"listen_address"`
	Ports         []ForwardPort `json:"ports"`
}

func (f *ForwardAddress) Get(ctx context.Context) (*ForwardAddressInfo, error) {
	_, metadata, err := request[forwardAddressMetadata](f.rest, ctx, http.MethodGet, f.path, nil)
	if err != nil {
		return nil, err
	}
	return &ForwardAddressInfo{
		ListenAddress: metadata.ListenAddress,
		Ports:         metadata.Ports,
	}, nil
}

type PutForwardAddressRequest struct {
	Ports []ForwardPort `json:"ports"`
}

func (f *ForwardAddress) Update(ctx context.Context, req PutForwardAddressRequest) error {
	_, err := f.rest.request(ctx, http.MethodPut, f.path, req)
	return err
}
