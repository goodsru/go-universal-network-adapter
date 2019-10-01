package main

import (
	"crypto/tls"
	"fmt"
	"github.com/goodsru/go-universal-network-adapter/models"
	"github.com/goodsru/go-universal-network-adapter/services"
	"io/ioutil"
)

func browseFtp() {
	adapter := services.NewUniversalNetworkAdapter()

	list, err := adapter.Browse(models.NewDestination("ftp://localhost:21", &models.Credentials{
		User:     "user",
		Password: "pass",
	}, nil))

	if err != nil {
		panic(err)
	}

	for _, item := range list {
		fmt.Println("name: %s, path: %s, size: %i", item.Name, item.Path, item.Size)
	}
}

func downloadFtp() {
	adapter := services.NewUniversalNetworkAdapter()

	remoteFile, err := models.NewRemoteFile(models.NewDestination("ftp://localhost:21/test1.txt", &models.Credentials{
		User:     "user",
		Password: "pass",
	}, nil))

	if err != nil {
		panic(err)
	}
	content, err := adapter.Download(remoteFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(content.Name)
	fmt.Println(content.Path)
	buf, err := ioutil.ReadAll(content.Blob)
	fmt.Println(string(buf))
}

func downloadFtpsTLSExplicit() {
	adapter := services.NewUniversalNetworkAdapter()

	public := `-----BEGIN CERTIFICATE-----\r\nsertData==\r\n-----END CERTIFICATE-----`
	private := `-----BEGIN RSA PRIVATE KEY-----\r\nsertData==\r\n-----END RSA PRIVATE KEY-----`
	keypair, err := tls.X509KeyPair([]byte(public), []byte(private))

	remoteFile, err := models.NewRemoteFile(models.NewDestination("ftps://localhost:21/test1.txt", &models.Credentials{
		User:      "user",
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{keypair}, InsecureSkipVerify: true},
		TLSMode:   models.TLSExplicit,
	}, nil))

	if err != nil {
		panic(err)
	}
	content, err := adapter.Download(remoteFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(content.Name)
	fmt.Println(content.Path)
	buf, err := ioutil.ReadAll(content.Blob)
	fmt.Println(string(buf))
}

func main() {
	browseFtp()
	downloadFtp()
	downloadFtpsTLSExplicit()
}
