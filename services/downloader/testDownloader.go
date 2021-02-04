package downloader

import "github.com/konart/go-universal-network-adapter/models"

type TestDownloader struct {
}

func (testDownloader *TestDownloader) Browse(destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	return []*models.RemoteFile{{
		Name: "test1.exe",
		Path: "/",
	}, {
		Name: "test2.jpg",
		Path: "/",
	}, {
		Name: "test3.txt",
		Path: "/",
	}}, nil
}

func (testDownloader *TestDownloader) Download(remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	return &models.RemoteFileContent{
		Name: remoteFile.Name,
		Path: "",
		Blob: &models.Blob{FilePath: "ee"},
	}, nil
}

func (testDownloader *TestDownloader) Stat(destination *models.ParsedDestination) (*models.RemoteFile, error) {
	return &models.RemoteFile{
		Name: "test1.exe",
		Path: "/",
	}, nil
}
