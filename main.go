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

	network, err := api.Network("lxdbr0")
	if err != nil {
		panic("Could not get network: " + err.Error())
	}
	adress, err := network.ForwardAddress("172.28.218.207")
	if err != nil {
		panic("Could not get forward addresses: " + err.Error())
	}
	info, err := adress.Get(context.Background())
	if err != nil {
		panic("Could not get forward address info: " + err.Error())
	}
	fmt.Printf("Forward addresses: %+v\n", info)
	err = adress.Update(context.Background(), lxd.PutForwardAddressRequest{
		Ports: []lxd.ForwardPort{
			{Protocol: "tcp", TargetPort: "25565", ListenPort: "25565", TargetAddr: "10.28.28.10"},
		},
	})
	if err != nil {
		panic("Could not update forward address: " + err.Error())
	}

	// instance, err := api.Instance(context.Background(), "wow")
	// if err != nil {
	// 	panic("Could not get instance: " + err.Error())
	// }
	// state, err := instance.GetState(context.Background())
	// if err != nil {
	// 	panic("Could not get instance state: " + err.Error())
	// }
	// fmt.Printf("Instance status: %+v\n", state)
}
