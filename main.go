package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"mcvds/internal/lxd"
	"mcvds/internal/mc"
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

	mcapi := &mc.PaperMCApi{
		BaseUrl: mc.PaperMCBaseUrl,
		Client:  http.DefaultClient,
	}

	builds, err := mcapi.GetBuilds(context.Background(), mc.PaperProject, "1.20.4")
	if err != nil || len(builds) == 0 {
		panic("Could not get PaperMC builds: " + err.Error())
	}

	res, err := http.Get(builds[0].Downloads["server:default"].URL)
	if err != nil {
		panic("Could not download PaperMC server jar: " + err.Error())
	}
	defer res.Body.Close()

	instance.PutFile(context.Background(), "/root/paper.jar", res.Body, &lxd.FileHeader{
		Mode: 0644,
		Uid:  0,
		Gid:  0,
	})
}
