package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/goodsru/go-universal-network-adapter/models"
	"github.com/goodsru/go-universal-network-adapter/services/downloader/http/integrationTest"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func Test_HttpDownloader_Unit(t *testing.T) {

}
func Test_HttpDownloader_IntegrationHttps(t *testing.T) {
	httpDownloader := &HttpDownloader{}
	fileName := "12345"
	data := `{"status": "ok"}`
	//success tests
	srv := integrationTest.StartWebServer("HTTPS")
	t.Run("Http_Download_ReturnsFileOverHttpsAndNoError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "https://localhost:9889/12345", Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		//set root certificate for access without authentication error
		caCert, err := ioutil.ReadFile("cert/rootCA1Cert.pem")
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
		client.Transport = tr
		result, err := httpDownloader.download(client, remoteFile)
		require.Nil(t, err, fmt.Sprintf("err == %v, Expect - nil", err))
		require.Equal(t, fileName, result.Name, fmt.Sprintf("Received file name %v, expected - %v", result.Name, fileName))

		_, err = os.Stat(result.Path)
		require.Nil(t, err)

		blobBytes, err := ioutil.ReadAll(result.Blob)
		require.Equal(t, data, string(blobBytes))
	})
	//error test
	t.Run("Http_Download_ReturnsFileOverHttpsAndError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "https://localhost:9889/12345", Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.NotNil(t, err, "Expect TLS handshake error from remote error: tls: bad certificate")
		require.Nil(t, result, "Expect empty result")
	})
	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}
func Test_HttpDownloader_IntegrationHttpsCertVerify(t *testing.T) {
	httpDownloader := &HttpDownloader{}
	fileName := "12345"
	data := `{"status": "ok"}`
	//success tests
	srv := integrationTest.StartWebServer("HTTPS_CERT_VERIFY")
	t.Run("Http_Download_ReturnsFileOverHttpsWithCertVerifyAndNoError", func(t *testing.T) {
		//load client certificate, signet by rootCA1
		cert, err := tls.LoadX509KeyPair("cert/ca1Client.pem", "cert/ca1Client.key")
		if err != nil {
			require.Nil(t, err)
		}
		// Create a CA certificate pool and add cert.pem to it
		caCert, err := ioutil.ReadFile("cert/rootCA1Cert.pem")
		if err != nil {
			require.Nil(t, err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "https://localhost:9890/12345", Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
		result, err := httpDownloader.download(client, remoteFile)
		require.Nil(t, err, fmt.Sprintf("err == %v, Expect - nil", err))
		require.Equal(t, fileName, result.Name, fmt.Sprintf("Received file name %v, expected - %v", result.Name, fileName))

		_, err = os.Stat(result.Path)
		require.Nil(t, err, fmt.Sprintf("err == %v, Expect - nil", err))

		blobBytes, err := ioutil.ReadAll(result.Blob)
		require.Equal(t, data, string(blobBytes))
	})
	// test with error
	t.Run("Http_Download_ReturnsFileOverHttpsWithCertVerifyAndAccessError", func(t *testing.T) {
		//load client certificate NOT signed by rootCA2
		cert, err := tls.LoadX509KeyPair("cert/ca2Client.pem", "cert/ca2Client.key")
		if err != nil {
			require.Nil(t, err)
		}
		// Create a CA certificate pool and add cert.pem to it
		caCert, err := ioutil.ReadFile("cert/rootCA2Cert.pem")
		if err != nil {
			require.Nil(t, err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "https://localhost:9890/12345", Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
		result, err := httpDownloader.download(client, remoteFile)
		require.NotNil(t, err, "Expect Err:x509.UnknownAuthorityError")
		require.Nil(t, result, "Expect empty result")
	})

	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}
func Test_HttpDownloader_IntegrationHttp(t *testing.T) {
	httpDownloader := &HttpDownloader{}
	fileName := "12345"
	data := `{"status": "ok"}`
	//success tests
	srv := integrationTest.StartWebServer("HTTP")
	t.Run("Http_Download_ReturnsFileOverHTTPAndNoError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "http://localhost:9888/12345", Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.Nil(t, err)
		require.NoError(t, err, fmt.Sprintf("err == %v, Expect - nil", err))
		require.Equal(t, fileName, result.Name, fmt.Sprintf("Received file name %v, expected - %v", result.Name, fileName))
		_, err = os.Stat(result.Path)
		require.Nil(t, err, fmt.Sprintf("err == %v, Expect - nil", err))
		blobBytes, err := ioutil.ReadAll(result.Blob)
		require.Equal(t, data, string(blobBytes))
	})
	t.Run("Http_Download_ReturnsFileOverHTTPWithBasicAuthAndNoError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "http://localhost:9888/basic/12345", Credentials: &models.Credentials{
			User:     `admin`,
			Password: `$CrazyUnforgettablePassword?`,
		}, Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.Nil(t, err, fmt.Sprintf("err == %v, Expect - nil", err))
		require.Equal(t, fileName, result.Name, fmt.Sprintf("Received file name %v, expected - %v", result.Name, fileName))
		_, err = os.Stat(result.Path)
		require.Nil(t, err, fmt.Sprintf("err == %v, Expect - nil", err))
		blobBytes, err := ioutil.ReadAll(result.Blob)
		require.Equal(t, data, string(blobBytes))
	})
	//error tests
	t.Run("Http_Download_ReturnsFileOverHTTPWithBasicAuthAndAuthError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "http://localhost:9888/basic/12345", Credentials: &models.Credentials{
			User:     `admin`,
			Password: `incorrectPass`,
		}, Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.NotNil(t, err, "Expect 401")
		require.Nil(t, result, "Expect empty result")
	})

	t.Run("Http_DownloadNonExistingFile_ReturnsError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "http://localhost:9888/noFile", Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.NotNil(t, err, "Expect 404")
		require.Nil(t, result, "Expect empty result")
	})
	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}
