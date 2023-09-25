package s3client

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Client struct {
	Bucket  string
	Path    string
	session *session.Session
}

func NewS3Client(s *session.Session) (*Client, error) {
	client := &Client{}
	client.Bucket = os.Getenv("ARTIFACTORY_BUCKET")
	client.Path = os.Getenv("ARTIFACTORY_META_WRITE_PATH")
	if client.Bucket == "" || client.Path == "" {
		return nil, fmt.Errorf("ARTIFACTORY_BUCKET or ARTIFACTORY_META_WRITE_PATH cannot be empty")
	}
	client.session = s
	return client, nil
}

func (c *Client) Write(data []byte) error {
	uploader := s3manager.NewUploader(c.session)
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(c.Path),
		Body:   bytes.NewReader(data),
	}
	out, err := uploader.Upload(upParams)
	fmt.Printf("Wrote metadata to %s\n", out.Location)
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	return nil
}
