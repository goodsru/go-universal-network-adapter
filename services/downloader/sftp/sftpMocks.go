package sftp

import (
	"io"
	"os"
	"time"

	"github.com/stretchr/testify/mock"
)

type fakeSftpClient struct {
	mock.Mock
}

func (ftp *fakeSftpClient) ReadDir(root string) ([]os.FileInfo, error) {
	args := ftp.Called(root)
	if args.Get(0) != nil {
		return args.Get(0).([]os.FileInfo), args.Error(1)
	}
	return ([]os.FileInfo)(nil), args.Error(1)
}

func (ftp *fakeSftpClient) Open(path string) (io.ReadCloser, error) {
	args := ftp.Called(path)
	if args.Get(0) != nil {
		return args.Get(0).(io.ReadCloser), args.Error(1)
	}
	return (io.ReadCloser)(nil), args.Error(1)

}

func (ftp *fakeSftpClient) Remove(path string) error {
	args := ftp.Called(path)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (ftp *fakeSftpClient) RemoveDirectory(path string) error {
	args := ftp.Called(path)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (ftp *fakeSftpClient) Close() error {
	args := ftp.Called()
	return args.Error(0)
}

func (ftp *fakeSftpClient) Stat(p string) (os.FileInfo, error) {
	args := ftp.Called(p)
	if args.Get(0) != nil {
		return args.Get(0).(os.FileInfo), args.Error(1)
	}
	return (os.FileInfo)(nil), args.Error(1)
}

type fakeFileInfo struct {
	mock.Mock
}

func (f *fakeFileInfo) Name() string {
	args := f.Called()
	return args.String(0)
}

func (f *fakeFileInfo) ModTime() time.Time {
	args := f.Called()
	return args.Get(0).(time.Time)
}
func (f *fakeFileInfo) IsDir() bool {
	args := f.Called()
	return args.Bool(0)
}
func (f *fakeFileInfo) Size() int64 {
	args := f.Called()
	return args.Get(0).(int64)
}
func (f *fakeFileInfo) Mode() os.FileMode {
	return 0777
}
func (f *fakeFileInfo) Sys() interface{} {
	return nil
}

type fakeReadCloser struct {
	data string
}

func (fakeReadCloser *fakeReadCloser) Read(p []byte) (n int, err error) {
	nn := copy(p, fakeReadCloser.data)
	return nn, io.EOF
}

func (fakeReadCloser *fakeReadCloser) Close() error {
	return nil
}
