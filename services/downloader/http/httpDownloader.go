//Package contains realization of class for file download via HTTP protocol
package http

import (
	"errors"
	"fmt"
	"github.com/goodsru/go-universal-network-adapter/models"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type HttpDownloader struct {
}

//Service method,that makes a HEAD request to remote server to get file size info
func (httpDownloader *HttpDownloader) Stat(destination *models.ParsedDestination) (*models.RemoteFile, error) {
	httpClient := httpDownloader.getClient(destination)
	return httpDownloader.stat(httpClient, destination)
}

//Not possible to implement this functionality thru HTTP protocol
func (httpDownloader *HttpDownloader) Browse(destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	return nil, fmt.Errorf("not implemented")
}

//Method allows download file from remote server, store it in temporary directory and
//return back RemoteFileContent with io.ReadCloser for further manipulations
func (httpDownloader *HttpDownloader) Download(remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	httpClient := httpDownloader.getClient(remoteFile.ParsedDestination)
	return httpDownloader.download(httpClient, remoteFile)
}

func (httpDownloader *HttpDownloader) download(client *http.Client, remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	localFile, err := ioutil.TempFile("", remoteFile.Name+".*")
	defer localFile.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("Get", remoteFile.ParsedDestination.Url, nil)
	if err != nil {
		return nil, err
	}
	destination := remoteFile.ParsedDestination
	user := destination.GetUser()
	password := destination.GetPassword()

	if password != "" {
		req.SetBasicAuth(user, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	_, err = io.Copy(localFile, resp.Body)

	return &models.RemoteFileContent{
		Name: remoteFile.Name,
		Path: localFile.Name(),
		Blob: &models.Blob{
			FilePath: localFile.Name(),
		},
	}, nil

}

//Return basic golang http Client with custom timeout from user request
func (httpDownloader *HttpDownloader) getClient(destination *models.ParsedDestination) *http.Client { //IHttpClient
	client := &http.Client{
		Timeout: destination.Timeout,
	}
	return client
}

func (httpDownloader *HttpDownloader) stat(client *http.Client, destination *models.ParsedDestination) (*models.RemoteFile, error) {
	req, err := http.NewRequest("HEAD", destination.Url, nil)
	if err != nil {
		return nil, err
	}
	user := destination.GetUser()
	password := destination.GetPassword()

	if password != "" {
		req.SetBasicAuth(user, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	fileSizeStr := resp.Header.Get("content-length")
	if len(fileSizeStr) == 0 {
		return nil, &models.UnaError{
			Code:    42,
			Message: "Не удалось получить размер файла",
		}
	}
	filePath := destination.GetPath()
	fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
	if err != nil {
		return nil, errors.New(resp.Status)
	}
	return &models.RemoteFile{Path: filePath, Size: fileSize, ParsedDestination: destination}, nil
}
