package ftp

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/fclairamb/ftpserver/server"
	"github.com/secsy/goftp"

	"github.com/bouk/monkey"
	"github.com/goodsru/go-universal-network-adapter/models"
	"github.com/goodsru/go-universal-network-adapter/tests/ftpServerDriver"
	assertLib "github.com/stretchr/testify/assert"
)

const (
	ServerIP = "127.0.0.1:2121"
)

type mockedFileInfo struct {
	os.FileInfo
	size    int64
	name    string
	modTime time.Time
}

func (m mockedFileInfo) Size() int64        { return m.size }
func (m mockedFileInfo) Name() string       { return m.name }
func (m mockedFileInfo) ModTime() time.Time { return m.modTime }

func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		return err
	}
	return nil
}

func TestFtpDownloaderUnit(t *testing.T) {
	assertions := assertLib.New(t)
	ftpDownloader := &FtpDownloader{}

	client := new(goftp.Client)

	t.Run("FtpMockedBrowseReturnsListAndNoError", func(t *testing.T) {
		monkey.UnpatchAll()
		monkey.Patch((*goftp.Client).ReadDir, func(client *goftp.Client, destination string) ([]os.FileInfo, error) {
			return []os.FileInfo{
				&mockedFileInfo{name: "1.jpg"},
				&mockedFileInfo{name: "2.jpg"},
				&mockedFileInfo{name: "3.jpg"},
				&mockedFileInfo{name: "4.jpg"},
			}, nil
		})

		parsedDest, _ := models.ParseDestination(&models.Destination{Url: "ftp://ftp.com/root"})

		list, err := ftpDownloader.browse(client, parsedDest)

		assertions.Nil(err, fmt.Sprintf("err == %v, expected - nil", err))
		assertions.Len(list, 4, fmt.Sprintf("found %v files, expected 4 files", len(list)))
		assertions.Equal(list[0].Name, "1.jpg", fmt.Sprintf("File name %v, expected %v", list[0].Name, "1.jpg"))
		assertions.Equal(list[3].Name, "4.jpg", fmt.Sprintf("File name %v, expected %v", list[0].Name, "4.jpg"))
	})

	t.Run("FtpMockedBrowseNonExistingFolderReturnsError", func(t *testing.T) {
		parsedDest, _ := models.ParseDestination(&models.Destination{Url: "ftp://ftp.com/dir123"})

		errText := "unknown directory"
		monkey.UnpatchAll()
		monkey.Patch((*goftp.Client).ReadDir, func(client *goftp.Client, destination string) ([]os.FileInfo, error) {
			return nil, fmt.Errorf(errText)
		})

		list, err := ftpDownloader.browse(client, parsedDest)

		assertions.NotNil(err)
		assertions.Nil(list)
		assertions.Equal(errText, err.Error())
	})

	t.Run("FtpMockedDownloadReturnsCorrectFile", func(t *testing.T) {
		fileName := "test.txt"
		expectedResult := "hello world"

		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "ftp://sftp.com:22/files/" + fileName})

		monkey.UnpatchAll()
		monkey.Patch(goftp.DialConfig, func(goftp.Config, ...string) (*goftp.Client, error) {
			return new(goftp.Client), nil
		})
		monkey.Patch((*goftp.Client).Retrieve, func(client *goftp.Client, filePath string, dest io.Writer) error {
			_, err := dest.Write([]byte(expectedResult))
			return err
		})

		fileContent, err := ftpDownloader.Download(remoteFile)

		if !assertions.NoError(err, fmt.Sprintf("err == %v, expected - nil", err)) {
			return
		}
		assertions.Equal(fileName, fileContent.Name, fmt.Sprintf("Received file name %v, expected - %v", fileContent.Name, fileName))

		_, err = os.Stat(fileContent.Path)
		assertions.Nil(err)
		blobBytes, err := ioutil.ReadAll(fileContent.Blob)
		assertions.Equal(expectedResult, string(blobBytes))

		err = fileContent.Blob.Close()
		assertions.NoError(err, fmt.Sprintf("err == %v, expected - nil", err))
	})

	t.Run("FtpMockedDownloadNonExistingFileReturnsError", func(t *testing.T) {
		fileName := "test.txt"
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "ftp://sftp.com:22/files/" + fileName})

		errText := "incorrect file"
		monkey.UnpatchAll()
		monkey.Patch((*goftp.Client).Retrieve, func(client *goftp.Client, filePath string, dest io.Writer) error {
			return fmt.Errorf(errText)
		})

		result, err := ftpDownloader.download(client, remoteFile)

		assertions.NotNil(err)
		assertions.Contains(err.Error(), errText)
		assertions.Nil(result)
	})
	monkey.UnpatchAll()
}

func Test_Locale_TLS(t *testing.T) {

	assertions := assertLib.New(t)

	fileName := "test.xml"
	fileToDeleteName := "test.xml"
	dirToDelete := "delete_dir"
	fileData := "hello world"
	tempFile := ""
	{
		err := CreateDirIfNotExist(ftpServerDriver.TempDir)
		if err != nil {
			t.Fatal("Couldn't create temp dir: "+ftpServerDriver.TempDir, err)
		}

		err = CreateDirIfNotExist(path.Join(ftpServerDriver.TempDir, dirToDelete))
		if err != nil {
			t.Fatal("Couldn't create temp dir: "+ftpServerDriver.TempDir, err)
		}

		localFile, err := ioutil.TempFile(ftpServerDriver.TempDir, fileName+".*")
		if err != nil {
			t.Fatal("Couldn't create temp file:", err)
		}
		tempFile = localFile.Name()

		_, err = localFile.Write([]byte(fileData))
		assertions.Nil(err)
		if err != nil {
			t.Fatal("Couldn't write to file:", err)
		}
		_ = localFile.Close()

		localFileToDelete, err := ioutil.TempFile(path.Join(ftpServerDriver.TempDir, dirToDelete), fileToDeleteName+".*")
		if err != nil {
			t.Fatal("Couldn't create temp file:", err)
		}
		tempFileToDelete := localFileToDelete.Name()

		_, err = localFileToDelete.Write([]byte(fileData))
		assertions.Nil(err)
		if err != nil {
			t.Fatal("Couldn't write to file:", err)
		}
		_ = localFileToDelete.Close()

		fileStat, err := os.Stat(tempFile)
		if err != nil {
			t.Fatal("Couldn't get stat from temp file:", err)
		}
		fileName = fileStat.Name()

		fileToDeleteStat, err := os.Stat(tempFileToDelete)
		if err != nil {
			t.Fatal("Couldn't get stat from temp file:", err)
		}
		fileToDeleteName = fileToDeleteStat.Name()

	}

	s := ftpServerDriver.NewTestServerWithDriver(&ftpServerDriver.ServerDriver{
		Debug:    false,
		TLS:      true,
		Settings: &server.Settings{ListenAddr: ServerIP},
	})
	defer s.Stop()

	ftpDownloader := FtpDownloader{}

	//BasicAuth/////////////////////////////////////////////////////////////////
	t.Run("FtpBrowseReturnsListAndNoErrorBasicAuth", func(t *testing.T) {
		parsedDest, err := models.ParseDestination(&models.Destination{
			Url: "ftp://" + ServerIP + "/",
			Credentials: &models.Credentials{
				User:     "test",
				Password: "test"},
		})

		list, err := ftpDownloader.Browse(parsedDest)

		assertions.NoError(err, fmt.Sprintf("err == %v, expected - nil", err))
		assertions.NotEqual(len(list), 0, "Files not found")
	})

	t.Run("FtpStatReturnsRemoteFileInfoAndNoErrorBasicAuth", func(t *testing.T) {
		parsedDest, err := models.ParseDestination(&models.Destination{
			Url: "ftp://" + ServerIP + "/" + fileName,
			Credentials: &models.Credentials{
				User:     "test",
				Password: "test"},
		})

		remoteFile, err := ftpDownloader.Stat(parsedDest)

		assertions.NoError(err, fmt.Sprintf("err == %v, expected - nil", err))
		assertions.Equal(remoteFile.Size, int64(len(fileData)), "File size is zero")
		assertions.Equal(remoteFile.Name, fileName, "File names must equal")
	})

	t.Run("FtpDownloadReturnsCorrectRemoteFileBasicAuth", func(t *testing.T) {
		remoteFile, _ := models.NewRemoteFile(&models.Destination{
			Url: "ftp://" + ServerIP + "/" + fileName,
			Credentials: &models.Credentials{
				User:     "test",
				Password: "test"},
		})

		result, err := ftpDownloader.Download(remoteFile)

		if err != nil {
			t.Fatal("err expected - nil", err)
		}
		assertions.Equal(fileName, result.Name, fmt.Sprintf("Received file name %v, expected - %v", result.Name, fileName))
	})

	t.Run("FtpBrowseBadCredentialsReturnsAuthError", func(t *testing.T) {
		parsedDest, err := models.ParseDestination(&models.Destination{
			Url: "ftp://" + ServerIP + "/",
			Credentials: &models.Credentials{
				User:     "test2",
				Password: "test"},
		})
		list, err := ftpDownloader.Browse(parsedDest)

		assertions.NotNil(err)
		assertions.Contains(err.Error(), "Authentication problem")
		assertions.Len(list, 0)
	})

	//DriverNotSupportDir
	//t.Run("FtpBrowseOnWrongDirReturnsError", func(t *testing.T) {
	//	parsedDest, err := models.ParseDestination(&models.Destination{
	//		Url: "ftp://" + ServerIP + "/bad_dir/some_some/some",
	//		Credentials: &models.Credentials{
	//			User:     "test",
	//			Password: "test"},
	//	})
	//
	//	list, err := ftpDownloader.Browse(parsedDest)
	//
	//	assertions.NotNil(err)
	//	assertions.Contains(err.Error(), "not exist")
	//	assertions.Len(list, 0)
	//})

	//t.Run("FtpGetClientTimeoutReturnsError", func(t *testing.T) {
	//	parsedDest, err := models.ParseDestination(&models.Destination{
	//		Url: "ftp://" + ServerIP + "/",
	//		Credentials: &models.Credentials{
	//			User:     "test",
	//			Password: "test"},
	//		Timeout: 1,
	//	})
	//	_, err = ftpDownloader.Browse(parsedDest)
	//
	//	//assertions.Nil(client)
	//	assertions.NotNil(err)
	//	assertions.Contains(err.Error(), "timeout")
	//})
	//BasicAuth/////////////////////////////////////////////////////////////////\

	//TLSExplicit/////////////////////////////////////////////////////////////////
	//t.Run("FtpBrowseReturnsCorrectListTLSExplicit", func(t *testing.T) {
	//	parsedDest, err := models.ParseDestination(&models.Destination{
	//		Url: "ftp://" + ServerIP + "/",
	//		Credentials: &models.Credentials{
	//			User:     "test",
	//			Password: "test",
	//			TLSConfig: &tls.Config{
	//				InsecureSkipVerify: true,
	//				ClientAuth:         tls.RequestClientCert,
	//				ClientSessionCache : tls.NewLRUClientSessionCache(0),
	//			},
	//			TLSMode: models.TLSExplicit,
	//		},
	//	})
	//
	//	list, err := ftpDownloader.Browse(parsedDest)
	//	assertions.NoError(err, fmt.Sprintf("err == %v, expected - nil", err))
	//	fileFound := false
	//	for _, remoteFile := range list {
	//		if fileFound = remoteFile.Name == fileName; fileFound {
	//			break
	//		}
	//	}
	//	assertions.True(fileFound, "File must be found")
	//})
	//
	////FtpDownload_ReturnsCorrectRemoteFile_TLSExplicit - old name
	//ftpDownloader = FtpDownloader{}
	//t.Run("FtpDownloadReturnsCorrectRemoteFileTLSExplicit", func(t *testing.T) {
	//	expectedResult := fileData
	//	remoteFile, _ := models.NewRemoteFile(&models.Destination{
	//		Url: "ftp://" + ServerIP + "/" + fileName,
	//		Credentials: &models.Credentials{
	//			User:     "test",
	//			Password: "test",
	//			TLSConfig: &tls.Config{
	//				InsecureSkipVerify: true,
	//				ClientAuth:         tls.RequestClientCert,
	//				ClientSessionCache : tls.NewLRUClientSessionCache(0),
	//			},
	//			TLSMode: models.TLSExplicit,
	//		},
	//	})
	//
	//	result, err := ftpDownloader.Download(remoteFile)
	//
	//	if !assertions.NoError(err, fmt.Sprintf("err == %v, expected - nil", err)) {
	//		return
	//	}
	//	assertions.Equal(fileName, result.Name, fmt.Sprintf("Received file name %v, expected - %v", result.Name, fileName))
	//
	//	blobBytes, err := ioutil.ReadAll(result.Blob)
	//	assertions.Equal(expectedResult, string(blobBytes))
	//
	//	err = result.Blob.Close()
	//	assertions.NoError(err, fmt.Sprintf("err == %v, expected - nil", err))
	//
	//})

	//TLSExplicit/////////////////////////////////////////////////////////////////\

	t.Run("FtpRemoveFileNoError", func(t *testing.T) {
		parsedDest, err := models.NewRemoteFile(&models.Destination{
			Url: "ftp://" + ServerIP + "/" + dirToDelete + "/" + fileToDeleteName,
			Credentials: &models.Credentials{
				User:     "test",
				Password: "test"},
		})
		assertions.Nil(err)
		err = ftpDownloader.Remove(parsedDest)
		assertions.Nil(err)
	})

	t.Run("FtpRemoveDirNoError", func(t *testing.T) {
		parsedDest, err := models.NewRemoteFile(&models.Destination{
			Url: "ftp://" + ServerIP + "/" + dirToDelete,
			Credentials: &models.Credentials{
				User:     "test",
				Password: "test"},
		})
		assertions.Nil(err)
		err = ftpDownloader.Remove(parsedDest)
		assertions.Nil(err)
	})
}
