package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"mcvds/internal/lxd"
	"net/http"
	"os"
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

	api := &lxd.Rest{
		Client:   initHttpClient(),
		Endpoint: endpoint,
	}

	d, err := api.Get(context.Background(), "/1.0/instances/present-hermit")
	if err != nil {
		fmt.Printf("Error getting instance state: %v\n", err)
		return
	}
	fmt.Printf("GetInstanceStateResponse: %+v\n", d.Metadata)

	// exec, err := api.ExecInstance(context.Background(), "present-hermit", lxd.ExecRequest{
	// 	Command:   []string{"/bin/df"},
	// 	WaitForWS: true,
	// })
	// if err != nil {
	// 	fmt.Printf("Error executing instance command: %v\n", err)
	// 	return
	// }

	// fmt.Printf("ExecInstanceResponse: %+v\n", exec)

	// dialer := websocket.Dialer{
	// 	TLSClientConfig: getTlsConfig(),
	// }

	// stdin, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, exec.Operation+"/websocket?secret="+exec.Metadata.Metadata.Fds["0"])
	// if err != nil {
	// 	fmt.Printf("Error connecting to WebSocket: %v\n", err)
	// 	return
	// }
	// defer stdin.Close()
	// stdout, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, exec.Operation+"/websocket?secret="+exec.Metadata.Metadata.Fds["1"])
	// if err != nil {
	// 	fmt.Printf("Error connecting to WebSocket: %v\n", err)
	// 	return
	// }
	// defer stdout.Close()
	// stderr, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, exec.Operation+"/websocket?secret="+exec.Metadata.Metadata.Fds["2"])
	// if err != nil {
	// 	fmt.Printf("Error connecting to WebSocket: %v\n", err)
	// 	return
	// }
	// defer stderr.Close()

	// for {
	// 	message, err := stdout.Read()
	// 	if err != nil {
	// 		fmt.Printf("Error reading from stdout WebSocket: %v\n", err)
	// 		break
	// 	}
	// 	fmt.Print(string(message))
	// }
}
