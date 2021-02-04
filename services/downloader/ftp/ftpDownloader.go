//go:generate mockery -case=underscore -name IFtpClient

package ftp

import (
	"github.com/konart/go-universal-network-adapter/models"
	"github.com/secsy/goftp"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type FtpDownloader struct {
}

type IFtpClient interface {
	ReadDir(destination string) ([]os.FileInfo, error)
	Retrieve(filePath string, dest io.Writer) error
}

func (ftpDownloader *FtpDownloader) Stat(destination *models.ParsedDestination) (*models.RemoteFile, error) {
	ftpClient, err := ftpDownloader.getClient(destination)
	if err != nil {
		return nil, err
	}
	defer ftpClient.Close()

	filePath := destination.GetPath()
	_, fileName := path.Split(filePath)

	remoteFiles, err := ftpDownloader.browse(ftpClient, destination)
	var remoteFile *models.RemoteFile
	for _, entry := range remoteFiles {
		if entry.Name == fileName {
			remoteFile = entry
			break
		}
	}
	return remoteFile, err
}

//Browse a files list in the server directory
func (ftpDownloader *FtpDownloader) Browse(destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	ftpClient, err := ftpDownloader.getClient(destination)
	if err != nil {
		return nil, err
	}

	defer ftpClient.Close()
	return ftpDownloader.browse(ftpClient, destination)
}

//Download the file. Get Blob io.ReadCloser
func (ftpDownloader *FtpDownloader) Download(remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	ftpClient, err := ftpDownloader.getClient(remoteFile.ParsedDestination)
	if err != nil {
		return nil, err
	}

	defer ftpClient.Close()

	return ftpDownloader.download(ftpClient, remoteFile)
}

func (ftpDownloader *FtpDownloader) download(ftpClient IFtpClient, remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	localFile, err := ioutil.TempFile("", remoteFile.Name+".*")
	if err != nil {
		return nil, err
	}
	defer localFile.Close()

	err = ftpClient.Retrieve(path.Join(remoteFile.Path, remoteFile.Name), localFile)
	if err != nil {
		return nil, err
	}

	return &models.RemoteFileContent{
		Name: remoteFile.Name,
		Path: localFile.Name(),
		Blob: &models.Blob{
			FilePath: localFile.Name(),
		},
	}, nil
}

//In credentials use only TLSConfig and TLSMode
func (ftpDownloader *FtpDownloader) getClient(destination *models.ParsedDestination) (*goftp.Client, error) {
	host := destination.GetHost()
	user := destination.GetUser()
	password := destination.GetPassword()

	config := goftp.Config{
		Timeout:   destination.Timeout,
		User:      user,
		Password:  password,
		TLSConfig: destination.Credentials.TLSConfig,
		TLSMode:   goftp.TLSMode(destination.Credentials.TLSMode),
	}
	client, err := goftp.DialConfig(config, host)

	if err != nil {
		return nil, err
	}

	return client, nil
}

func (ftpDownloader *FtpDownloader) browse(client IFtpClient, destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	entryList, err := client.ReadDir(destination.GetPath())
	result := make([]*models.RemoteFile, 0)
	if err != nil {
		return nil, err
	}
	for _, entry := range entryList {
		result = append(result, &models.RemoteFile{Name: entry.Name(), Path: destination.GetPath(), Size: entry.Size(),
			ParsedDestination: destination, Lastmod: entry.ModTime()})
	}
	return result, nil
}

//type FtpResponse interface {
//	Read(p []byte) (n int, err error)
//	Close() error
//}
//
//type ftpClientWrapper struct {
//	*ftp.ServerConn
//}
//
//func (ftpClientWrapper *ftpClientWrapper) Retr(filePath string) (FtpResponse, error) {
//	return ftpClientWrapper.ServerConn.Retr(filePath)
//}
