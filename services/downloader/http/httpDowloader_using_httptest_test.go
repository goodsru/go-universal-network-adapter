package http

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"github.com/goodsru/go-universal-network-adapter/models"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func hasher(s string) []byte {
	val := sha256.Sum256([]byte(s))
	return val[:]
}
func authHandler(handler http.HandlerFunc, userhash, passhash []byte, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare(hasher(user),
			userhash) != 1 || subtle.ConstantTimeCompare(hasher(pass), passhash) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}
func handlers() http.Handler {
	userhash := hasher("admin")
	passhash := hasher("$CrazyUnforgettablePassword?")
	realm := "Please enter username and password"
	r := http.NewServeMux()
	r.HandleFunc("/12345", indexHandler)
	r.HandleFunc("/basic/12345", authHandler(indexHandler, userhash, passhash, realm))

	return r
}
func Test_HttpDownloader_UsingHttpTest(t *testing.T) {
	httpDownloader := &HttpDownloader{}
	fileName := "12345"
	data := `{"status": "ok"}`
	//success tests

	ts := httptest.NewServer(handlers())
	defer ts.Close()
	t.Run("Http_Download_ReturnsFileOverHTTPAndNoError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: ts.URL + "/12345", Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.Nil(t, err)
		require.NoError(t, err, fmt.Sprintf("err == %v, ожидается - nil", err))
		require.Equal(t, fileName, result.Name, fmt.Sprintf("Получено имя файла %v, ожидается - %v", result.Name, fileName))
		_, err = os.Stat(result.Path)
		require.Nil(t, err, fmt.Sprintf("err == %v, ожидается - nil", err))
		blobBytes, err := ioutil.ReadAll(result.Blob)
		require.Equal(t, data, string(blobBytes))
	})
	t.Run("Http_Download_ReturnsFileOverHTTPWithBasicAuthAndNoError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: ts.URL + "/basic/12345", Credentials: &models.Credentials{
			User:     `admin`,
			Password: `$CrazyUnforgettablePassword?`,
		}, Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.Nil(t, err)
		require.NoError(t, err, fmt.Sprintf("err == %v, ожидается - nil", err))
		require.Equal(t, fileName, result.Name, fmt.Sprintf("Получено имя файла %v, ожидается - %v", result.Name, fileName))
		_, err = os.Stat(result.Path)
		require.Nil(t, err, fmt.Sprintf("err == %v, ожидается - nil", err))
		blobBytes, err := ioutil.ReadAll(result.Blob)
		require.Equal(t, data, string(blobBytes))
	})
	//error tests
	t.Run("Http_Download_ReturnsFileOverHTTPWithBasicAuthAndAuthError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: ts.URL + "/basic/12345", Credentials: &models.Credentials{
			User:     `admin`,
			Password: `incorrectPass`,
		}, Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.NotNil(t, err, "Ожидается 401")
		require.Nil(t, result, "Ожидается пустой результат")
	})

	t.Run("Http_DownloadNonExistingFile_ReturnsError", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: ts.URL + "/noFile", Timeout: 3 * time.Minute})
		client := httpDownloader.getClient(remoteFile.ParsedDestination)
		result, err := httpDownloader.download(client, remoteFile)
		require.NotNil(t, err, "Ожидается 404")
		require.Nil(t, result, "Ожидается пустой результат")
	})
}
