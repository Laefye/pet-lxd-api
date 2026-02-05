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
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCAs,
	}
	return tlsConfig
}

func initHttpClient() http.Client {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: getTlsConfig(),
		},
	}
	return client
}

const serverURL = "127.0.0.1:8443"

func main() {
	endpoint := lxd.Endpoint(serverURL)

	// api := &lxd.Rest{
	// 	Client: initHttpClient(),
	// 	Host:   serverURL,
	// }

	dialer := websocket.Dialer{
		TLSClientConfig: getTlsConfig(),
	}

	conn, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, "/1.0/events")
	if err != nil {
		fmt.Printf("Error connecting to WebSocket: %v\n", err)
		return
	}
	defer conn.Conn.Close()

	fmt.Println("Connected to WebSocket, waiting for events...")

	for {
		message, err := conn.Read()
		if err != nil {
			fmt.Printf("Error reading from WebSocket: %v\n", err)
			break
		}
		fmt.Printf("Received event: %s\n", string(message))
	}
}
