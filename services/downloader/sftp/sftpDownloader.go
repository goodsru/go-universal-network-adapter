// sftp file downloader implementation
package sftp

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"time"

	"github.com/goodsru/go-universal-network-adapter/models"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// sftp file downloader implementation
type SftpDownloader struct {
}

func (sftpDownloader *SftpDownloader) Stat(destination *models.ParsedDestination) (*models.RemoteFile, error) {
	sftpClient, err := sftpDownloader.getClient(destination)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()
	return sftpDownloader.stat(sftpClient, destination)
}

func (sftpDownloader *SftpDownloader) Browse(destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	sftpClient, err := sftpDownloader.getClient(destination)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()
	return sftpDownloader.browse(sftpClient, destination)
}

func (sftpDownloader *SftpDownloader) Download(remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	sftpClient, err := sftpDownloader.getClient(remoteFile.ParsedDestination)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()
	return sftpDownloader.download(sftpClient, remoteFile)
}

func (sftpDownloader *SftpDownloader) Remove(remoteFile *models.RemoteFile) error {
	sftpClient, err := sftpDownloader.getClient(remoteFile.ParsedDestination)
	if err != nil {
		return err
	}
	defer sftpClient.Close()
	return sftpDownloader.remove(sftpClient, remoteFile)
}

func (sftpDownloader *SftpDownloader) stat(client iSftpClient, destination *models.ParsedDestination) (*models.RemoteFile, error) {
	filePath := destination.GetPath()
	stat, err := client.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("destination is a directory")
	}

	return &models.RemoteFile{Name: stat.Name(), Path: filePath, Size: stat.Size(), Lastmod: stat.ModTime(), IsDir: stat.IsDir(), ParsedDestination: destination}, nil
}

func (sftpDownloader *SftpDownloader) browse(client iSftpClient, destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	folderPath := destination.GetPath()
	result := make([]*models.RemoteFile, 0)

	items, err := client.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		result = append(result, &models.RemoteFile{Name: item.Name(), Path: folderPath, Size: item.Size(), Lastmod: item.ModTime(), IsDir: item.IsDir(), ParsedDestination: destination})
	}

	return result, nil
}

func (sftpDownloader *SftpDownloader) download(sftpClient iSftpClient, remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	ftpFile, err := sftpClient.Open(path.Join(remoteFile.Path, remoteFile.Name))
	if err != nil {
		return nil, err
	}

	localFile, err := ioutil.TempFile("", remoteFile.Name+".*")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(localFile, ftpFile)
	if err != nil {
		return nil, err
	}

	localFile.Close()
	ftpFile.Close()

	return &models.RemoteFileContent{
		Name: remoteFile.Name,
		Path: localFile.Name(),
		Blob: &models.Blob{
			FilePath: localFile.Name(),
		},
	}, nil
}

func (sftpDownloader *SftpDownloader) getClient(destination *models.ParsedDestination) (iSftpClient, error) {
	user := destination.GetUser()
	pass := destination.GetPassword()

	var auth []ssh.AuthMethod

	if pass != "" {
		auth = append(auth, ssh.Password(pass))
	}

	if signer, err := destination.GetRsaPrivateKey(); err == nil {
		auth = append(auth, ssh.PublicKeys(signer))
	}

	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         destination.Timeout,
	}

	url := destination.GetHost()
	client, err := sftpDownloader.sshDialWithTimeout("tcp", url, sshConfig, sshConfig.Timeout)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}

	return &sftpClientWrapper{sftpClient}, nil
}

// ssh.Dial can hang during authentication, the 'timeout' being set in the config only applying to establishment of the initial connection.
// This function is effectively ssh.Dial with the ability to set a deadline on the underlying connection.
// https://github.com/Yeba/fuchsia-tool/commit/652b0acfd634aea432eb2432dcc8ea5a37dccc3b?diff=split
func (sftpDownloader *SftpDownloader) sshDialWithTimeout(network, addr string, config *ssh.ClientConfig, hardTimeout time.Duration) (*ssh.Client, error) {
	conn, err := net.DialTimeout(network, addr, config.Timeout)
	if err != nil {
		return nil, err
	}
	if err := conn.SetDeadline(time.Now().Add(hardTimeout)); err != nil {
		conn.Close()
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		c.Close()
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func (sftpDownloader *SftpDownloader) remove(client iSftpClient, remoteFile *models.RemoteFile) error {
	filePath := path.Join(remoteFile.Path, remoteFile.Name)
	stat, err := client.Stat(filePath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return client.RemoveDirectory(filePath)
	}
	return client.Remove(filePath)
}

type iSftpClient interface {
	ReadDir(root string) ([]os.FileInfo, error)
	Open(path string) (io.ReadCloser, error)
	Stat(p string) (os.FileInfo, error)
	Remove(path string) error
	RemoveDirectory(path string) error
	Close() error
}

type sftpClientWrapper struct {
	*sftp.Client
}

//cast sftp.File to io.ReadCloser for testing purposes
func (wrap *sftpClientWrapper) Open(path string) (io.ReadCloser, error) {
	return wrap.Client.Open(path)
}
