package models

import (
	"io"
	"os"
	"path"
	"time"
)

// Remote file model
type RemoteFile struct {
	// filename
	Name string
	// filepath
	Path string
	// remote file parsed destination model
	ParsedDestination *ParsedDestination `json:"-"`
	// file size, bytes
	Size int64
	// file modification date
	Lastmod time.Time
	IsDir   bool
}

// Constructor for RemoteFile
func NewRemoteFile(destination *Destination) (*RemoteFile, error) {
	parsedDest, err := ParseDestination(destination)
	if err != nil {
		return nil, err
	}

	filePath := parsedDest.GetPath()
	dir, file := path.Split(filePath)

	return &RemoteFile{ParsedDestination: parsedDest, Path: dir, Name: file}, nil
}

// Downloaded remote file content model
type RemoteFileContent struct {
	// filename
	Name string
	// filepath
	Path string
	// io.ReadCloser to read remote file content
	Blob io.ReadCloser
}

// Wrapper around remote file, implementing io.ReadCloser
type Blob struct {
	// путь до файла
	FilePath string
	// файл
	file *os.File
}

func (blob *Blob) Read(p []byte) (n int, err error) {
	if blob.file == nil {
		file1, err := os.Open(blob.FilePath)
		blob.file = file1
		if err != nil {
			return 0, err
		}
	}

	return blob.file.Read(p)
}

func (blob *Blob) Close() error {
	if blob.file != nil {
		err := blob.file.Close()
		if err != nil {
			return err
		}
		err = os.Remove(blob.FilePath)
		if err != nil {
			return err
		}
		blob.file = nil
	}
	return nil
}
