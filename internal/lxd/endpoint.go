package lxd

type Endpoint string

func (e Endpoint) Https(path string) string {
	return "https://" + string(e) + path
}

func (e Endpoint) Wss(path string) string {
	return "wss://" + string(e) + path
}
