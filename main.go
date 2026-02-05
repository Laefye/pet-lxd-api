package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"mcvds/internal/lxd"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Api struct {
	client http.Client
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

func initHttpClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: getTlsConfig(),
		},
	}
	return client
}

const serverURL = "172.28.218.207:8443"

func main() {
	endpoint := lxd.Endpoint(serverURL)

	api := &lxd.Rest{
		Client:   initHttpClient(),
		Endpoint: endpoint,
	}

	res, exec, err := lxd.R[lxd.ExecMetadata](api, context.Background(), http.MethodPost, lxd.MustParsePath("/1.0/instances/wow/exec"), lxd.ExecRequest{
		Command:   []string{"/usr/sbin/chpasswd", "gamesrv"},
		WaitForWS: true,
	})
	if err != nil {
		panic(err)
	}

	dialer := websocket.Dialer{
		TLSClientConfig: getTlsConfig(),
	}
	stdin, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, lxd.MustParsePath(res.Operation).Join("websocket").WithSecret(exec.Metadata.Fds["0"]))
	if err != nil {
		panic(err)
	}
	stdout, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, lxd.MustParsePath(res.Operation).Join("websocket").WithSecret(exec.Metadata.Fds["1"]))
	if err != nil {
		panic(err)
	}
	defer stdout.Close()
	stderr, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, lxd.MustParsePath(res.Operation).Join("websocket").WithSecret(exec.Metadata.Fds["2"]))
	if err != nil {
		panic(err)
	}
	defer stderr.Close()

	stdin.Write([]byte("root:password\n"))
	stdin.Close()

	message, err := stdout.ReadMessage()
	if err != nil {
		panic(err)
	}
	println("STDOUT:", string(message))
}
