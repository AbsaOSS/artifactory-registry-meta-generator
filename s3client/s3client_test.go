package s3client

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

const respMsg = `<?xml version="1.0" encoding="UTF-8"?>
<CompleteUploadOutput>
   <Location></Location>
   <Bucket>mockValue</Bucket>
   <Key>mockValue</Key>
   <ETag>mockValue</ETag>
</CompleteUploadOutput>`

func contains(src []string, s string) bool {
	for _, v := range src {
		if s == v {
			return true
		}
	}
	return false
}

func TestUpload(t *testing.T) {
	s := unit.Session
	s.Handlers.Send.Clear()
	params := []interface{}{}
	names := []string{}
	ignoreOps := []string{}
	var m sync.Mutex
	s.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		if !contains(ignoreOps, r.Operation.Name) {
			names = append(names, r.Operation.Name)
			params = append(params, r.Params)
		}

		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(respMsg))),
		}
		switch data := r.Data.(type) {
		case *s3.PutObjectOutput:
			data.VersionId = aws.String("VERSION-ID")
			data.ETag = aws.String("ETAG")
		}
	})
	c := &Client{
		Path:    "/foo",
		Bucket:  "test-bucket",
		session: s,
	}
	data := []byte(`{"/docker/registry/path": "shasum"}`)
	err := c.Write(data)
	assert.NoError(t, err)

}
