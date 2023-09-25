package main

import (
	"encoding/json"
	"fmt"

	"github.com/AbsaOSS/artifactory-registry-meta-generator/artifactory"
	"github.com/AbsaOSS/artifactory-registry-meta-generator/s3client"

	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {

	af, err := artifactory.CreateArtifactory()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	s3client, err := s3client.NewS3Client(session.Must(session.NewSession()))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	paths, err := af.GetFilePaths()
	if err != nil {
		fmt.Printf("Error getting paths %s\n", err)
		return
	}
	data := make(map[string]string)

	for _, f := range paths {
		generateMeta(f, data)
	}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Failed to marshal json")
		return
	}
	err = s3client.Write(jsonStr)
	if err != nil {
		fmt.Printf("Failed to upload metadata to s3: %s\n", err)
	}
}
