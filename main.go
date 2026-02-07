package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
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

	instance, err := api.Instance(context.Background(), "wow")
	if err != nil {
		panic("Could not get instance: " + err.Error())
	}
	state, err := instance.GetState(context.Background())
	if err != nil {
		panic("Could not get instance state: " + err.Error())
	}
	println("Instance status:", state.Status)

	file, err := instance.GetFile(context.Background(), "/home/owo")
	if err != nil {
		var apiErr *lxd.LxdApiError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
			fmt.Println("File not found")
			return
		}
		panic("Could not get instance files: " + err.Error())
	}
	defer file.Close()
	fmt.Printf("File info - %+v\n", file.Header())
	if file.IsDir() {
		fmt.Println("Instance files:")
		for _, f := range file.FileList() {
			fmt.Println(" -", f)
		}
	} else {
		content, err := io.ReadAll(file.GetReader())
		if err != nil {
			panic("Could not read file content: " + err.Error())
		}
		fmt.Println("File content:\n", string(content))
	}

	testFile := bytes.NewBufferString("Hello World\n")
	err = instance.PutFile(context.Background(), "/home/owo/test.txt", testFile, &lxd.FileHeader{Uid: 1000, Gid: 1000, Mode: 0644})
	if err != nil {
		panic("Could not put file: " + err.Error())
	}

}
