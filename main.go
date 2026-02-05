package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"mcvds/internal/lxd"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Api struct {
	client *http.Client
	url    string
}

func getTlsConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
	if err != nil {
		panic("Could not load client key pair: " + err.Error())
	}

	rootCAs := x509.NewCertPool()
	serverCert, err := os.ReadFile("server.crt")
	if err != nil {
		panic("Could not read server certificate: " + err.Error())
	}
	rootCAs.AppendCertsFromPEM(serverCert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            rootCAs,
		InsecureSkipVerify: true,
	}
	return tlsConfig
}

const serverURL = "172.28.218.207:8443"

func main() {
	endpoint := lxd.Endpoint(serverURL)

	api := &lxd.Rest{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: getTlsConfig(),
			},
		},
		Dialer: &websocket.Dialer{
			TLSClientConfig: getTlsConfig(),
		},
		Endpoint: endpoint,
	}

	fd, err := api.Exec(context.Background(), lxd.MustParsePath("/1.0/instances/wow/exec"), lxd.ExecRequest{
		Command:   []string{"/usr/sbin/chpasswd"},
		WaitForWS: true,
	})
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	fd.Stdin.Write([]byte("owo:password\n"))
	fd.Stdin.Close()

	out, err := fd.Stdout.ReadMessage()
	if err != nil && err != io.EOF {
		panic(err)
	}
	println(string(out))
}
