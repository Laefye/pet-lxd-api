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

	query := lxd.Query{
		Project: "test",
	}

	instanceCreatingTask, err := api.Request(context.Background(), http.MethodPost, "/1.0/instances"+query.String(), lxd.CreateInstanceRequest{
		Source: lxd.InstanceSource{
			Alias: "leafos",
			Type:  "image",
		},
		Type:  lxd.InstanceTypeVM,
		Start: true,
	})
	if err != nil {
		fmt.Printf("Error creating instance: %v\n", err)
		return
	}

	fmt.Printf("CreateInstanceResponse: %+v\n", instanceCreatingTask)

	instanceCreated, err := api.Request(context.Background(), http.MethodGet, instanceCreatingTask.Operation+"/wait", nil)
	if err != nil {
		fmt.Printf("Error waiting for instance creation: %v\n", err)
		return
	}

	fmt.Printf("InstanceCreatedResponse: %+v\n", instanceCreated.Metadata)

	path := lxd.ParsePath(instanceCreated.Metadata.Resources.Instances[0])

	for {
		state, err := api.Request(context.Background(), http.MethodGet, path.Join("state").String(), nil)
		if err != nil {
			fmt.Printf("Error getting instance state: %v\n", err)
			return
		}
		fmt.Printf("InstanceStateResponse: %+v\n", state.Metadata)
		if state.Metadata.Processes > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Instance is started and ready for exec")

	exec, err := api.Request(context.Background(), http.MethodPost, path.Join("exec").String(), lxd.ExecRequest{
		Command:   []string{"/usr/bin/wget", "https://fill-data.papermc.io/v1/objects/2617fbbe4a9c0642ee5e0176a459b64992c7a308a1773e9bd42ef1d2d7bec25a/paper-1.21.11-105.jar", "-O", "/root/server.jar"},
		WaitForWS: true,
	})
	if err != nil {
		fmt.Printf("Error executing instance command: %v\n", err)
		return
	}

	fmt.Printf("ExecInstanceResponse: %+v\n", exec)

	dialer := websocket.Dialer{
		TLSClientConfig: getTlsConfig(),
	}

	stdin, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, exec.Operation+"/websocket?secret="+exec.Metadata.Metadata.Fds["0"])
	if err != nil {
		fmt.Printf("Error connecting to WebSocket: %v\n", err)
		return
	}
	defer stdin.Close()
	stdout, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, exec.Operation+"/websocket?secret="+exec.Metadata.Metadata.Fds["1"])
	if err != nil {
		fmt.Printf("Error connecting to WebSocket: %v\n", err)
		return
	}
	defer stdout.Close()
	stderr, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, exec.Operation+"/websocket?secret="+exec.Metadata.Metadata.Fds["2"])
	if err != nil {
		fmt.Printf("Error connecting to WebSocket: %v\n", err)
		return
	}
	defer stderr.Close()

	for {
		message, err := stderr.Read()
		if err != nil {
			fmt.Printf("Error reading from stdout WebSocket: %v\n", err)
			break
		}
		fmt.Print(string(message))
	}
}
