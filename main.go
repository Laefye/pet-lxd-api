package main

import (
	"crypto/tls"
	"crypto/x509"
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
	// endpoint := lxd.Endpoint(serverURL)

	// api := &lxd.Rest{
	// 	Client:   initHttpClient(),
	// 	Endpoint: endpoint,
	// }

	// instances := lxd.ParsePath("/1.0/instances").WithProject("test")

	// instanceCreatingTask, err := api.Request(context.Background(), http.MethodPost, instances.String(), lxd.CreateInstanceRequest{
	// 	Source: lxd.InstanceSource{
	// 		Alias: "leafos",
	// 		Type:  "image",
	// 	},
	// 	Type:  lxd.InstanceTypeVM,
	// 	Start: true,
	// })
	// if err != nil {
	// 	fmt.Printf("Error creating instance: %v\n", err)
	// 	return
	// }

	// fmt.Printf("CreateInstanceResponse: %+v\n", instanceCreatingTask)

	// instanceCreated, err := api.Request(context.Background(), http.MethodGet, lxd.ParsePath(instanceCreatingTask.Operation).Join("wait").String(), nil)
	// if err != nil {
	// 	fmt.Printf("Error waiting for instance creation: %v\n", err)
	// 	return
	// }

	// fmt.Printf("InstanceCreatedResponse: %+v\n", instanceCreated.Metadata)

	// metadata := lxd.RestMetadata{}
	// err = json.Unmarshal(instanceCreated.Metadata, &metadata)
	// if err != nil {
	// 	fmt.Printf("Error unmarshaling instance metadata: %v\n", err)
	// 	return
	// }
	// path := lxd.ParsePath(metadata.Resources.Instances[0])

	// for {
	// 	state, err := api.Request(context.Background(), http.MethodGet, path.Join("state").String(), nil)
	// 	if err != nil {
	// 		fmt.Printf("Error getting instance state: %v\n", err)
	// 		return
	// 	}
	// 	metadata := lxd.RestMetadata{}
	// 	err = json.Unmarshal(state.Metadata, &metadata)
	// 	if err != nil {
	// 		fmt.Printf("Error unmarshaling instance state metadata: %v\n", err)
	// 		return
	// 	}
	// 	fmt.Printf("InstanceStateResponse: %+v\n", metadata)
	// 	if metadata.Processes > 0 {
	// 		break
	// 	}
	// 	time.Sleep(1 * time.Second)
	// }

	// fmt.Println("Instance is started and ready for exec")

	// exec, err := api.Request(context.Background(), http.MethodPost, path.Join("exec").String(), lxd.ExecRequest{
	// 	Command:   []string{"/usr/bin/wget", "https://fill-data.papermc.io/v1/objects/2617fbbe4a9c0642ee5e0176a459b64992c7a308a1773e9bd42ef1d2d7bec25a/paper-1.21.11-105.jar", "-O", "/root/server.jar"},
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

	// execOperation := lxd.ParsePath(exec.Operation)
	// metadata = lxd.RestMetadata{}
	// err = json.Unmarshal(exec.Metadata, &metadata)
	// if err != nil {
	// 	fmt.Printf("Error unmarshaling exec metadata: %v\n", err)
	// 	return
	// }

	// stdin, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, execOperation.Join("websocket").WithSecret(metadata.Metadata.Fds["0"]).String())
	// if err != nil {
	// 	fmt.Printf("Error connecting to WebSocket: %v\n", err)
	// 	return
	// }
	// defer stdin.Close()
	// stdout, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, execOperation.Join("websocket").WithSecret(metadata.Metadata.Fds["1"]).String())
	// if err != nil {
	// 	fmt.Printf("Error connecting to WebSocket: %v\n", err)
	// 	return
	// }
	// defer stdout.Close()
	// stderr, err := lxd.ConnectWebsocket(context.Background(), dialer, endpoint, execOperation.Join("websocket").WithSecret(metadata.Metadata.Fds["2"]).String())
	// if err != nil {
	// 	fmt.Printf("Error connecting to WebSocket: %v\n", err)
	// 	return
	// }
	// defer stderr.Close()

	// for {
	// 	message, err := stderr.Read()
	// 	if err != nil {
	// 		fmt.Printf("Error reading from stdout WebSocket: %v\n", err)
	// 		break
	// 	}
	// 	fmt.Print(string(message))
	// }
}
