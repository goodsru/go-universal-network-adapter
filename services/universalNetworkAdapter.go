package services

import (
	"fmt"

	"github.com/goodsru/go-universal-network-adapter/contracts"
	"github.com/goodsru/go-universal-network-adapter/models"
	"github.com/goodsru/go-universal-network-adapter/services/downloader/ftp"
	"github.com/goodsru/go-universal-network-adapter/services/downloader/http"
	"github.com/goodsru/go-universal-network-adapter/services/downloader/s3"
	"github.com/goodsru/go-universal-network-adapter/services/downloader/sftp"
)

const (
	HTTP  = "http"
	HTTPS = "https"
	FTP   = "ftp"
	FTPS  = "ftps"
	SFTP  = "sftp"
	S3    = "s3"
)

type UniversalNetworkAdapter struct {
	downloaderMap map[string]contracts.Downloader
}

func NewUniversalNetworkAdapter() *UniversalNetworkAdapter {
	adapter := &UniversalNetworkAdapter{downloaderMap: make(map[string]contracts.Downloader)}

	httpDownloader := &http.HttpDownloader{}
	adapter.RegisterDownloader(httpDownloader, HTTP)
	adapter.RegisterDownloader(httpDownloader, HTTPS)

	adapter.RegisterDownloader(&ftp.FtpDownloader{}, FTP)
	adapter.RegisterDownloader(&ftp.FtpDownloader{}, FTPS)
	adapter.RegisterDownloader(&sftp.SftpDownloader{}, SFTP)

	adapter.RegisterDownloader(&s3.S3Downloader{}, S3)

	return adapter
}

func (adapter *UniversalNetworkAdapter) Stat(destination *models.Destination) (*models.RemoteFile, error) {
	parsedDestination, err := models.ParseDestination(destination)
	if err != nil {
		return nil, err
	}
	downloader, err := adapter.getDownloader(parsedDestination)
	if err != nil {
		return nil, err
	}
	return downloader.Stat(parsedDestination)
}

func (adapter *UniversalNetworkAdapter) Browse(destination *models.Destination) ([]*models.RemoteFile, error) {
	parsedDestination, err := models.ParseDestination(destination)
	if err != nil {
		return nil, err
	}
	downloader, err := adapter.getDownloader(parsedDestination)
	if err != nil {
		return nil, err
	}
	return downloader.Browse(parsedDestination)
}

func (adapter *UniversalNetworkAdapter) Download(remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	downloader, err := adapter.getDownloader(remoteFile.ParsedDestination)
	if err != nil {
		return nil, err
	}
	return downloader.Download(remoteFile)
}

func (adapter *UniversalNetworkAdapter) Remove(remoteFile *models.RemoteFile) error {
	downloader, err := adapter.getDownloader(remoteFile.ParsedDestination)
	if err != nil {
		return err
	}
	return downloader.Remove(remoteFile)
}

func (adapter *UniversalNetworkAdapter) getDownloader(parsedDestination *models.ParsedDestination) (contracts.Downloader, error) {
	var scheme string

	if parsedDestination.Protocol != "" {
		scheme = parsedDestination.Protocol
	} else {
		scheme = parsedDestination.GetScheme()
	}

	downloader, ok := adapter.downloaderMap[scheme]
	if !ok {
		return nil, fmt.Errorf("не найден загрузчик для URL: %v, Scheme: %s", parsedDestination.Url, scheme)
	}
	return downloader, nil
}

func (adapter *UniversalNetworkAdapter) RegisterDownloader(downloader contracts.Downloader, scheme string) {
	if _, ok := adapter.downloaderMap[scheme]; ok == true {
		fmt.Printf("Будет произведена замена существующего загрузчика для схемы %v\n", scheme)
	}
	adapter.downloaderMap[scheme] = downloader
}
