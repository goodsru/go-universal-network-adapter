package s3

import (
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/goodsru/go-universal-network-adapter/models"
	// "goods.ru/go-universal-network-adapter/models"
)

type S3Downloader struct{}

// переписать
func (s *S3Downloader) Stat(destination *models.ParsedDestination) (*models.RemoteFile, error) {
	client, err := s.getClient(destination)
	if err != nil {
		return nil, err
	}

	arr := strings.SplitAfterN(destination.GetPath(), "/", 2)
	path := arr[1]

	files, err := s.browse(client, destination)
	if err != nil {
		return nil, err
	}

	var remoteFile *models.RemoteFile
	for _, f := range files {
		if f.Name == path {
			remoteFile = f
			break
		}
	}

	return remoteFile, nil
}

func (s *S3Downloader) Browse(destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	client, err := s.getClient(destination)
	if err != nil {
		return nil, err
	}

	return s.browse(client, destination)
}

func (s *S3Downloader) Download(file *models.RemoteFile) (*models.RemoteFileContent, error) {
	client, err := s.getClient(file.ParsedDestination)
	if err != nil {
		return nil, err
	}

	return s.download(client, file)
}

func (s *S3Downloader) Remove(remoteFile *models.RemoteFile) error {
	return nil
}

func (s *S3Downloader) getClient(destination *models.ParsedDestination) (*s3.S3, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(destination.GetUser(), destination.GetPassword(), ""),
		Endpoint:         aws.String(destination.GetHost()),
		Region:           aws.String("ru-central1"), // временный костыль
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	return svc, nil
}

func (s *S3Downloader) download(client *s3.S3, remoteFile *models.RemoteFile) (*models.RemoteFileContent, error) {
	localFile, err := ioutil.TempFile("", remoteFile.Name+".*")
	if err != nil {
		return nil, err
	}

	defer localFile.Close()

	in := s3.GetObjectInput{
		Bucket: aws.String(strings.TrimPrefix(remoteFile.ParsedDestination.GetPath(), "/")),
		Key:    aws.String(remoteFile.Name),
	}

	dm := s3manager.NewDownloaderWithClient(client)
	_, err = dm.Download(localFile, &in)
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

func (s *S3Downloader) browse(client *s3.S3, destination *models.ParsedDestination) ([]*models.RemoteFile, error) {
	in := s3.ListObjectsInput{
		Bucket: aws.String(strings.TrimPrefix(destination.GetPath(), "/")),
	}

	out, err := client.ListObjects(&in)
	if err != nil {
		return nil, err
	}

	files := make([]*models.RemoteFile, len(out.Contents))
	for i, o := range out.Contents {
		file := &models.RemoteFile{
			Name:              *o.Key,
			Path:              *o.Key,
			ParsedDestination: destination,
			Size:              *o.Size,
			Lastmod:           *o.LastModified,
		}
		files[i] = file
	}

	return files, nil
}
