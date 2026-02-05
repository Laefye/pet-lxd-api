package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"mcvds/internal/lxd"
	"mcvds/internal/mc"
	"net/http"
	"os"
	"strings"
	"sync"
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

func initHttpClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: getTlsConfig(),
		},
	}
	return client
}

const serverURL = "127.0.0.1:8443"

const paper = "https://fill.papermc.io"

var recommendFlags = []string{
	"-XX:+AlwaysPreTouch",
	"-XX:+DisableExplicitGC",
	"-XX:+ParallelRefProcEnabled",
	"-XX:+PerfDisableSharedMem",
	"-XX:+UnlockExperimentalVMOptions",
	"-XX:+UseG1GC",
	"-XX:G1HeapRegionSize=8M",
	"-XX:G1HeapWastePercent=5",
	"-XX:G1MaxNewSizePercent=40",
	"-XX:G1MixedGCCountTarget=4",
	"-XX:G1MixedGCLiveThresholdPercent=90",
	"-XX:G1NewSizePercent=30",
	"-XX:G1RSetUpdatingPauseTimePercent=5",
	"-XX:G1ReservePercent=20",
	"-XX:InitiatingHeapOccupancyPercent=15",
	"-XX:MaxGCPauseMillis=200",
	"-XX:MaxTenuringThreshold=1",
	"-XX:SurvivorRatio=30",
}

func createInstance(ctx context.Context, api *lxd.Rest) (*lxd.Path, error) {
	response, err := api.Request(ctx, http.MethodPost, lxd.MustParsePath("/1.0/instances"), &lxd.CreateInstanceRequest{
		Source: lxd.InstanceSource{
			Alias: "leafos",
			Type:  "image",
		},
		Type:  lxd.InstanceTypeVM,
		Start: true,
		Config: map[string]string{
			"limits.cpu":    "2",
			"limits.memory": "4GiB",
		},
	})
	if err != nil {
		return nil, err
	}
	_, instanceCreatingTask, err := lxd.R[lxd.RestMetadata](api, ctx, http.MethodGet, lxd.MustParsePath(response.Operation).Join("wait"), nil)
	if err != nil {
		return nil, err
	}
	path := lxd.MustParsePath(instanceCreatingTask.Resources.Instances[0])
	return &path, nil
}

func waitInstanceReady(ctx context.Context, api *lxd.Rest, path lxd.Path) error {
	path = path.Join("state")
	for {
		_, metadata, err := lxd.R[lxd.RestMetadata](api, ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		fmt.Printf("InstanceStateResponse: %+v\n", metadata)
		if metadata.Processes > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func executeCommand(ctx context.Context, api *lxd.Rest, path lxd.Path, command []string, request *lxd.ExecRequest) (*lxd.Response, *lxd.RestMetadata, error) {
	if request == nil {
		request = &lxd.ExecRequest{
			Command:   command,
			WaitForWS: true,
		}
	}
	response, metadata, err := lxd.R[lxd.RestMetadata](api, ctx, http.MethodPost, path.Join("exec"), request)
	if err != nil {
		return nil, nil, err
	}
	return response, metadata, nil
}

func toStdout(name string, reader *lxd.WebSocketStream) error {
	for {
		bytes, err := reader.ReadMessage()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Printf("[%s] %s", name, string(bytes))
	}
	return nil
}

func readFds(ctx context.Context, dialer websocket.Dialer, endpoint lxd.Endpoint, operation lxd.Path, metadata *lxd.RestMetadata) error {
	stdin, err := lxd.ConnectWebsocket(ctx, dialer, endpoint, operation.Join("websocket").WithSecret(metadata.Metadata.Fds["0"]))
	if err != nil {
		return err
	}
	defer stdin.Close()
	stdout, err := lxd.ConnectWebsocket(ctx, dialer, endpoint, operation.Join("websocket").WithSecret(metadata.Metadata.Fds["1"]))
	if err != nil {
		return err
	}
	defer stdout.Close()
	stderr, err := lxd.ConnectWebsocket(ctx, dialer, endpoint, operation.Join("websocket").WithSecret(metadata.Metadata.Fds["2"]))
	if err != nil {
		return err
	}
	defer stderr.Close()

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, stdout)
		fmt.Println("STDOUT closed")
		stdin.Close()
	}()

	go func() {
		defer wg.Done()
		io.Copy(os.Stderr, stderr)
		fmt.Println("STDERR closed")
	}()

	go func() {
		defer wg.Done()
		io.Copy(stdin, os.Stdin)
		fmt.Println("STDIN closed")
	}()

	wg.Wait()
	return nil
}

func execAndWaitCommand(ctx context.Context, api *lxd.Rest, dialer websocket.Dialer, endpoint lxd.Endpoint, path lxd.Path, command []string, request *lxd.ExecRequest) error {
	execResponse, execMetadata, err := executeCommand(ctx, api, path, command, request)
	if err != nil {
		return err
	}
	fmt.Printf("Created exec operation at path: %s\n", execResponse.Operation)
	return readFds(ctx, dialer, endpoint, lxd.MustParsePath(execResponse.Operation), execMetadata)
}

func uploadEula(ctx context.Context, api *lxd.Rest, path lxd.Path, filename string) error {
	url := path.Join("files").WithQuery("path", filename)
	buffer := bytes.NewBufferString("eula=true")
	resp, err := api.Upload(ctx, url, buffer, http.Header{
		"Content-Type": []string{"application/octet-stream"},
		"X-LXD-uid":    []string{"1000"},
		"X-LXD-gid":    []string{"1000"},
	})
	if err != nil {
		return err
	}
	log.Printf("Upload response: %+v\n", resp)
	return nil
}

func main() {
	paper := mc.PaperMCApi{
		BaseUrl: paper,
		Client:  http.Client{},
	}

	builds, err := paper.GetBuilds(context.Background(), "paper", "1.21.11")
	if err != nil {
		panic(err)
	}

	fmt.Printf("PaperMC Latest Build: %+v\n", builds[0].ID)

	endpoint := lxd.Endpoint(serverURL)

	api := &lxd.Rest{
		Client:   initHttpClient(),
		Endpoint: endpoint,
	}

	instancePath, err := createInstance(context.Background(), api)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created instance at path: %s\n", instancePath.String())

	err = waitInstanceReady(context.Background(), api, *instancePath)
	if err != nil {
		panic(err)
	}

	fmt.Println("Instance is started and ready for exec")

	dialer := websocket.Dialer{
		TLSClientConfig: getTlsConfig(),
	}

	err = execAndWaitCommand(context.Background(), api, dialer, endpoint, *instancePath, []string{"/usr/sbin/adduser", "gamesrv", "-D"}, nil)
	if err != nil {
		panic(err)
	}

	err = execAndWaitCommand(context.Background(), api, dialer, endpoint, *instancePath, []string{}, &lxd.ExecRequest{
		Command:   []string{"/usr/bin/wget", builds[0].Downloads["server:default"].URL, "-O", "/home/gamesrv/server.jar"},
		Cwd:       "/home/gamesrv",
		User:      1000,
		Group:     1000,
		WaitForWS: true,
	})
	if err != nil {
		panic(err)
	}

	err = uploadEula(context.Background(), api, *instancePath, "/home/gamesrv/eula.txt")
	if err != nil {
		panic(err)
	}

	java := fmt.Sprintf("java %s -jar server.jar nogui", strings.Join(recommendFlags, " "))

	fmt.Printf("Starting server with command: %s\n", java)

	suWithJava := []string{"/bin/su", "-", "gamesrv", "-c", java}

	err = execAndWaitCommand(context.Background(), api, dialer, endpoint, *instancePath, []string{}, &lxd.ExecRequest{
		Command:   suWithJava,
		WaitForWS: true,
		Cwd:       "/home/gamesrv",
	})
	if err != nil {
		panic(err)
	}
}
