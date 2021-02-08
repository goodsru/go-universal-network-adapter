package tests

import (
	"github.com/goodsru/go-universal-network-adapter/models"
	"github.com/goodsru/go-universal-network-adapter/services"
	"github.com/goodsru/go-universal-network-adapter/services/downloader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"strings"
	"testing"
	"time"
)

func TestUniversalNetworkAdapter_TestDownloader(t *testing.T) {
	t.Run("BrowseTestSchemeWithoutRegistration_ReturnsError", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		_, err := adapter.Browse(&models.Destination{Url: "test://goods.ru"})
		if err == nil {
			t.Errorf("err = nil, ожидается - не nil")
		}
	})

	t.Run("RegisterAndBrowse_ReturnsNoError", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		adapter.RegisterDownloader(&downloader.TestDownloader{}, "test")
		_, err := adapter.Browse(&models.Destination{Url: "test://goods.ru"})
		if err != nil {
			t.Errorf("err != nil, ожидается - nil")
		}
	})

	t.Run("RegisterAndDownload_ReturnsExpectedFilename", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		adapter.RegisterDownloader(&downloader.TestDownloader{}, "test")
		fileName := "file.txt"
		remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "test://goods.ru/" + fileName})
		file, err := adapter.Download(remoteFile)
		if err != nil {
			t.Errorf("err != nil, ожидается - nil")
			return
		}
		if file.Name != fileName {
			t.Errorf("имя загруженного файла %v, ожидается - %v", file.Name, fileName)
		}
	})
}

//func TestUniversalNetworkAdapter_HttpDownloader(t *testing.T) {
//	t.Run("BrowseHttp_IsPreconfiguredButNotImplemented", func(t *testing.T) {
//		adapter := services.NewUniversalNetworkAdapter()
//		_, err := adapter.Browse(&models.Destination{Url: "http://goods.ru"})
//		if err == nil {
//			t.Errorf("err = nil, ожидается - не nil")
//			return
//		}
//		expectedError := "not implemented"
//		if err.Error() != expectedError {
//			t.Errorf("err = %v, ожидается - %v", err.Error(), expectedError)
//		}
//	})
//
//	t.Run("DownloadHttp_ReturnsCorrectRemoteFile", func(t *testing.T) {
//		adapter := services.NewUniversalNetworkAdapter()
//		fileName := "logo_goods.svg"
//		result, err := adapter.Download(&models.RemoteFile{Name: fileName, Path: "https://goods.ru/generated/img/logo_goods.svg"})
//		if err != nil {
//			t.Errorf("err == %v, ожидается - nil", err.Error())
//			return
//		}
//		if result.Name != fileName {
//			t.Errorf("Получено имя файла %v, ожидается - %v", result.Name, fileName)
//		}
//	})
//}

type MockFtpDownloader struct {
	mock.Mock
}

func (m *MockFtpDownloader) Browse(destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	args := m.Called(destination)
	return args.Get(0).([]*models.RemoteFile), args.Error(1)
}

func (m *MockFtpDownloader) Stat(destination *models.ParsedDestination) (*models.RemoteFile, error) {
	args := m.Called(destination)
	return args.Get(0).(*models.RemoteFile), args.Error(1)
}

func (m *MockFtpDownloader) Download(remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	args := m.Called(remoteFile)
	return args.Get(0).(*models.RemoteFileContent), args.Error(1)
}

type testBlob struct {
	reader io.Reader
}

func (blob *testBlob) Close() error {
	return nil
}

func (blob *testBlob) Read(p []byte) (n int, err error) {
	return blob.reader.Read(p)
}

func createReaderFromString(data string) io.ReadCloser {
	reader := strings.NewReader(data)
	blob := &testBlob{reader: reader}
	return blob
}

func TestUniversalNetworkAdapter_FtpDownloader(t *testing.T) {

	destinationList := models.NewDestination("ftp://xyml_exchange:3F&:Z-jghw9-S%5C%5C&b)S6U-=@sftp.goods.ru:22/files/stocks/1120", nil, nil)
	parsedDestination, _ := models.ParseDestination(destinationList)
	browsResponse := []*models.RemoteFile{
		{Name: "test1.json", Path: "testPath", Size: 60, Lastmod: time.Date(2019, 04, 01, 20, 0, 0, 0, time.UTC)},
		{Name: "test2.json", Path: "testPath", Size: 60, Lastmod: time.Date(2019, 04, 01, 20, 0, 1, 0, time.UTC)},
	}

	t.Run("BrowseFtp_ReturnsList", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		mockFtpDownloader := &MockFtpDownloader{}
		mockFtpDownloader.On("Browse", parsedDestination).Return(browsResponse, nil)

		adapter.RegisterDownloader(mockFtpDownloader, "ftp")

		list, err := adapter.Browse(destinationList)

		assert.Nil(t, err, "err ожидается - nil")
		assert.Equal(t, 2, len(list), "Ожидается наличие 2 файлов")
		assert.Equal(t, "test1.json", list[0].Name, "Ожидаем файл test1.json")
	})

	remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "ftp://xyml_exchange:3F&:Z-jghw9-S%5C%5C&b)S6U-=@sftp.goods.ru:22/files/stocks/1120/" + "test1.json"})
	textData := "Test data"
	blob := createReaderFromString(textData)
	downloadResponse := &models.RemoteFileContent{Name: "test1.json", Path: "testPath", Blob: blob}

	t.Run("DownloadFtp_ReturnsCorrectRemoteFile", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		mockFtpDownloader := &MockFtpDownloader{}
		mockFtpDownloader.On("Download", remoteFile).Return(downloadResponse, nil)

		adapter.RegisterDownloader(mockFtpDownloader, "ftp")

		remoteFileContent, err := adapter.Download(remoteFile)

		assert.Nil(t, err, "err ожидается - nil")
		assert.Equal(t, "test1.json", remoteFileContent.Name, "Ожидаем файл test1.json")
		slice := make([]byte, len(textData))
		_, err = remoteFileContent.Blob.Read(slice)
		assert.Nil(t, err, "err ожидается - nil")

		assert.Equal(t, textData, string(slice), "Ожидаем файл test1.json")
		err = remoteFileContent.Blob.Close()
		assert.Nil(t, err, "err ожидается - nil")

	})

}

func TestUniversalNetworkAdapter_sFtpDownloader(t *testing.T) {

	destinationList := models.NewDestination("sftp://xyml_exchange:3F&:Z-jghw9-S%5C%5C&b)S6U-=@sftp.goods.ru:22/files/stocks/1120", nil, nil)
	parsedDestination, _ := models.ParseDestination(destinationList)
	browsResponse := []*models.RemoteFile{
		{Name: "test1.json", Path: "testPath", Size: 60, Lastmod: time.Date(2019, 04, 01, 20, 0, 0, 0, time.UTC)},
		{Name: "test2.json", Path: "testPath", Size: 60, Lastmod: time.Date(2019, 04, 01, 20, 0, 0, 0, time.UTC)},
	}

	t.Run("BrowseFtp_ReturnsList", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		mockFtpDownloader := &MockFtpDownloader{}
		mockFtpDownloader.On("Browse", parsedDestination).Return(browsResponse, nil)

		adapter.RegisterDownloader(mockFtpDownloader, "sftp")

		list, err := adapter.Browse(destinationList)

		assert.Nil(t, err, "err ожидается - nil")
		assert.Equal(t, 2, len(list), "Ожидается наличие 2 файлов")
		assert.Equal(t, "test1.json", list[0].Name, "Ожидаем файл test1.json")
	})

	remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "sftp://xyml_exchange:3F&:Z-jghw9-S%5C%5C&b)S6U-=@sftp.goods.ru:22/files/stocks/1120/" + "test1.json"})
	textData := "Test data"
	blob := createReaderFromString(textData)
	downloadResponse := &models.RemoteFileContent{Name: "test1.json", Path: "testPath", Blob: blob}

	t.Run("DownloadFtp_ReturnsCorrectRemoteFile", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		mockFtpDownloader := &MockFtpDownloader{}
		mockFtpDownloader.On("Download", remoteFile).Return(downloadResponse, nil)

		adapter.RegisterDownloader(mockFtpDownloader, "sftp")

		remoteFileContent, err := adapter.Download(remoteFile)

		assert.Nil(t, err, "err ожидается - nil")
		assert.Equal(t, "test1.json", remoteFileContent.Name, "Ожидаем файл test1.json")
		slice := make([]byte, len(textData))
		_, err = remoteFileContent.Blob.Read(slice)
		assert.Nil(t, err, "err ожидается - nil")

		assert.Equal(t, textData, string(slice), "Ожидаем файл test1.json")
		err = remoteFileContent.Blob.Close()
		assert.Nil(t, err, "err ожидается - nil")

	})

}

type MockHttpDownloader struct {
	mock.Mock
}

func (m *MockHttpDownloader) Browse(destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	args := m.Called(destination)
	return args.Get(0).([]*models.RemoteFile), args.Error(1)
}

func (m *MockHttpDownloader) Stat(destination *models.ParsedDestination) (*models.RemoteFile, error) {
	args := m.Called(destination)
	return args.Get(0).(*models.RemoteFile), args.Error(1)
}

func (m *MockHttpDownloader) Download(remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	args := m.Called(remoteFile)
	return args.Get(0).(*models.RemoteFileContent), args.Error(1)
}

func TestUniversalNetworkAdapter_HTTPDownloader(t *testing.T) {

	remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "http://goods.ru/generated/img/" + "logo_goods.svg"})

	textData := "Test data"
	blob := createReaderFromString(textData)
	downloadResponse := &models.RemoteFileContent{Name: "test1.json", Path: "testPath", Blob: blob}

	t.Run("DownloadHTTP_ReturnsCorrectRemoteFile", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		mockHttpDownloader := &MockHttpDownloader{}
		mockHttpDownloader.On("Download", remoteFile).Return(downloadResponse, nil)

		adapter.RegisterDownloader(mockHttpDownloader, "http")

		remoteFileContent, err := adapter.Download(remoteFile)

		assert.Nil(t, err, "err ожидается - nil")
		assert.Equal(t, "test1.json", remoteFileContent.Name, "Ожидаем файл test1.json")
		slice := make([]byte, len(textData))
		_, err = remoteFileContent.Blob.Read(slice)
		assert.Nil(t, err, "err ожидается - nil")

		assert.Equal(t, textData, string(slice), "Ожидаем файл test1.json")
		err = remoteFileContent.Blob.Close()
		assert.Nil(t, err, "err ожидается - nil")

	})

}

func TestUniversalNetworkAdapter_HTTPSDownloader(t *testing.T) {

	remoteFile, _ := models.NewRemoteFile(&models.Destination{Url: "https://goods.ru/generated/img/" + "logo_goods.svg"})
	textData := "Test data"
	blob := createReaderFromString(textData)
	downloadResponse := &models.RemoteFileContent{Name: "test1.json", Path: "testPath", Blob: blob}

	t.Run("DownloadHTTP_ReturnsCorrectRemoteFile", func(t *testing.T) {
		adapter := services.NewUniversalNetworkAdapter()
		mockHttpDownloader := &MockHttpDownloader{}
		mockHttpDownloader.On("Download", remoteFile).Return(downloadResponse, nil)

		adapter.RegisterDownloader(mockHttpDownloader, "https")

		remoteFileContent, err := adapter.Download(remoteFile)

		assert.Nil(t, err, "err ожидается - nil")
		assert.Equal(t, "test1.json", remoteFileContent.Name, "Ожидаем файл test1.json")
		slice := make([]byte, len(textData))
		_, err = remoteFileContent.Blob.Read(slice)
		assert.Nil(t, err, "err ожидается - nil")

		assert.Equal(t, textData, string(slice), "Ожидаем файл test1.json")
		err = remoteFileContent.Blob.Close()
		assert.Nil(t, err, "err ожидается - nil")

	})

}
