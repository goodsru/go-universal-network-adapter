package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/konart/go-universal-network-adapter/models"
	"github.com/konart/go-universal-network-adapter/services"
	"io/ioutil"
	"time"
)

var defaultHttpTimeout = time.Duration(int64(time.Minute))

func downloadHttp() {
	adapter := services.NewUniversalNetworkAdapter()

	remoteFile, err := models.NewRemoteFile(models.NewDestination("https://golangcode.com/images/avatar.jpg", nil, &defaultHttpTimeout))

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
func downloadHttpBasicAuth() {
	adapter := services.NewUniversalNetworkAdapter()
	remoteFile, err := models.NewRemoteFile(models.NewDestination("http://localhost:443/avatar.jpg", &models.Credentials{
		User:     `admin`,
		Password: `$CrazyUnforgettablePassword?`,
	}, &defaultHttpTimeout))

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
func downloadHttpCertificateAuth() {
	adapter := services.NewUniversalNetworkAdapter()

	//load client certificate, signet by rootCA1
	//in certFile parameter need to setup your local public key
	cert, err := tls.LoadX509KeyPair("cert/ca1Client.pem", "cert/ca1Client.key")
	if err != nil {
		panic(err)
	}
	// Create a CA certificate pool and add cert.pem to it
	//in file name need to setup path to a root CA certificate
	caCert, err := ioutil.ReadFile("cert/rootCA1Cert.pem")
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	remoteFile, err := models.NewRemoteFile(models.NewDestination("http://localhost:443/avatar.jpg", &models.Credentials{
		User:      "user",
		TLSConfig: tlsConfig,
	}, &defaultHttpTimeout))

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
	downloadHttp()
	downloadHttpBasicAuth()
	downloadHttpCertificateAuth()
}
