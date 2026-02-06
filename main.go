package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"mcvds/internal/lxd"
	"net/http"
	"os"
	"time"

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

	instance, err := api.CreateInstance(context.Background(), lxd.InstanceCreationRequest{
		Source: lxd.InstanceSource{Type: "image", Alias: "leafos"},
		Start:  true,
		Type:   lxd.InstanceTypeVM,
	})
	if err != nil {
		panic(err)
	}
	for {
		time.Sleep(5 * time.Second)
		state, err := instance.State(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Printf("Instance state: %+v\n", state)
		if state.Processes > 0 {
			break
		}
	}
	fmt.Printf("Created instance at path: %+v\n", instance)
}
