package lxd

import "context"

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

type ForwardAddress struct {
	ListenAddress string        `json:"listen_address"`
	Ports         []ForwardPort `json:"ports"`
}

type Network struct {
	rest *Rest
	path resourcePath
}

func (r *Rest) Network(ctx context.Context, name string) (*Network, error) {
	path := r.base.join("networks").withoutQuery("project").join(name)
	return &Network{rest: r, path: path}, nil
}
