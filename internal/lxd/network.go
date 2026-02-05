package lxd

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
