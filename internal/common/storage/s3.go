package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/twofas/2fas-server/internal/common/logging"
)

type S3 struct {
	sess *session.Session
}

func NewS3Storage(sess *session.Session) *S3 {
	return &S3{
		sess: sess,
	}
}

func (s *S3) Get(path string) (file *os.File, err error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	downloader := s3manager.NewDownloader(s.sess)

	f, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(directory),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download the object from s3: %w", err)
	}

	return f, nil
}

func (s *S3) Save(path string, data io.Reader) (location string, err error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	uploader := s3manager.NewUploader(s.sess)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(directory),
		Key:    aws.String(name),
		Body:   data,
	})

	if err != nil {
		logging.WithFields(logging.Fields{
			"error":    err.Error(),
			"bucket":   directory,
			"filename": name,
		}).Error("Cannot upload file")

		return "", err
	}

	return result.Location, nil
}

func (s *S3) Move(oldPath, newPath string) (location string, err error) {
	sourceDirectory := filepath.Dir(oldPath)
	sourceName := filepath.Base(oldPath)

	svc := s3.New(s.sess)

	file, err := s.Get(oldPath)

	if err != nil {
		return "", err
	}

	newLocation, err := s.Save(newPath, file)

	if err != nil {
		return "", err
	}

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(sourceDirectory),
		Key:    aws.String(sourceName)},
	)

	if err != nil {
		logging.WithFields(logging.Fields{
			"error":    err.Error(),
			"bucket":   sourceDirectory,
			"filename": sourceName,
		}).Error("Cannot delete file")

		return newLocation, err
	}

	return newLocation, nil
}
