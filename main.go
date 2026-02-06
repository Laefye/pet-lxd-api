package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
	api := lxd.NewRest(
		serverURL,
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: getTlsConfig(),
			},
		},
		&websocket.Dialer{
			TLSClientConfig: getTlsConfig(),
		},
	)

	instance, err := api.GetInstance(context.Background(), "actual-mongrele")
	if err != nil {
		panic("Could not get instance: " + err.Error())
	}
	state, err := instance.GetState(context.Background())
	if err != nil {
		panic("Could not get instance state: " + err.Error())
	}
	fmt.Printf("Instance state: %s, processes: %d\n", state.Status, state.Processes)
	websockets, err := instance.Exec(context.Background(), lxd.ExecRequest{
		Command:   []string{"ls", "/"},
		WaitForWS: true,
	})
	if err != nil {
		panic("Could not execute command: " + err.Error())
	}
	defer websockets.Close()
	for {
		message, err := websockets["1"].ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				panic("Websocket closed unexpectedly: " + err.Error())
			}
			break
		}
		fmt.Printf("Received message: %s\n", string(message))
	}
}
