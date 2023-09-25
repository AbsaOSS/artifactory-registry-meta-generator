package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Paths struct {
	Paths []Path `json:"children"`
}

type Path struct {
	Folder bool   `json:"folder"`
	Uri    string `json:"uri"`
}

type File struct {
	Checksums Checksum `json:"checksums"`
}

type Checksum struct {
	SHA1   string `json:"sha1"`
	MD5    string `json:"md5"`
	SHA256 string `json:"sha256"`
}

type Artifactory struct {
	URL  string
	User string
	Pass string
}

func IteratePath(artifactory Artifactory, p string) ([]string, error) {
	body, err := getData(artifactory.URL+p, artifactory.User, artifactory.Pass)
	if err != nil {
		return nil, err
	}
	list := Paths{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, path := range list.Paths {
		if !path.Folder {
			paths = append(paths, []string{artifactory.URL + p + path.Uri}...)
			continue
		}

		p, err := IteratePath(artifactory, p+path.Uri)
		if err != nil {
			return nil, err
		}
		paths = append(paths, p...)
	}
	return paths, nil
}

const (
	dockerBaseURL = "/docker/registry/v2"
)

func getData(uri, u, p string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(u, p)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	//b, _ := ioutil.ReadAll(resp.Body)

	//fmt.Printf("SSDA %+v\n", string(b))
	return io.ReadAll(resp.Body)

}

func check(e error) {
	if e != nil {
		fmt.Printf("Error occured: %s\n", e.Error())
		panic(e)
	}
}

// /bks-docker-local/stefanprodan/podinfo/5.2.0/manifest.json ->
// /bks-docker-local/stefanprodan/podinfo/_manifests/tags/5.2.0/current/link
func normalizePath(s string) string {
	return strings.TrimPrefix(s, os.Getenv("ARTIFACTORY_URL"))
}

func isManifest(s string) bool {
	return strings.HasSuffix(s, "manifest.json")
}

func isLayer(s string) bool {
	return strings.Contains(s, "sha256__")
}

// /bks-docker-local/cert-manager-controller/_manifests/tags/v0.12.0-venafi/current/link
// /bks-docker-local/cert-manager-controller/_manifests/revisions/sha256/044c3ca8c12c47635ecf137e6132ea615b4a65b5d540a3796332ac00724c2541/link
func convertLinkToRev(p string, sha256 string) string {
	reg := regexp.MustCompile(`/tags.+`)
	return dockerBaseURL + string(reg.ReplaceAllString(p, fmt.Sprintf("/revisions/sha256/%s/link", sha256)))
}

// /docker/registry/v2/repositories/bks-docker-local/cert-manager-controller/_manifests/revisions/sha256/044c3ca8c12c47635ecf137e6132ea615b4a65b5d540a3796332ac00724c2541/link
// /docker/registry/v2/blobs/sha256/04/044c3ca8c12c47635ecf137e6132ea615b4a65b5d540a3796332ac00724c2541/data
func getLinkToBlob(file Checksum, data map[string]string) {
	shaPath := sha256path(file.SHA256)
	data[fmt.Sprintf("/blobs/sha256/%s/data", shaPath)] = file.SHA1
}
func sha256path(s string) string {
	return fmt.Sprintf("%s/%s", s[0:2], s)
}

// /bks-docker-local/cert-manager-controller/v0.12.0-venafi/sha256__c5c9eab06e7db9e76641b4fe8351725d5d3d40100db3f0efaa411807022441e6 ->
// /repositories/bks-docker-local/cert-manager-controller/_layers/sha256/b02a7525f878e61fc1ef8a7405a2cc17f866e8de222c1c98fd6681aff6e509db/link
func handleLayerLink(s string, sha256 string) string {
	parts := strings.Split(s, "/")
	repo := parts[1]
	image := parts[2]
	return dockerBaseURL + fmt.Sprintf("/repositories/%s/%s/_layers/sha256/%s/link", repo, image, sha256)
}

func generateManifest(s string, info Checksum, data map[string]string) {
	pathParts := strings.Split(s, "/")
	partsLen := len(pathParts)
	suffix := pathParts[partsLen-2]
	prefix := strings.Join(pathParts[0:partsLen-2], "/")
	// link withou sha256
	link := fmt.Sprintf("/repositories%s/_manifests/tags/%s/current/link", prefix, suffix)
	revLink := convertLinkToRev(link, info.SHA256)
	data[dockerBaseURL+link] = info.SHA256
	data[revLink] = info.SHA256
	getLinkToBlob(info, data)
}
func generateLayer(s string, info Checksum, data map[string]string) {
	data[handleLayerLink(s, info.SHA256)] = info.SHA256
	getLinkToBlob(info, data)
}
func generateMeta(p string, sum Checksum, data map[string]string) {
	p = normalizePath(p)
	if isManifest(p) {
		generateManifest(p, sum, data)
	}
	if isLayer(p) {
		generateLayer(p, sum, data)
	}

}

func writeToS3(bucket, path string, data []byte) error {
	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	}
	_, err := uploader.Upload(upParams)
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	return nil
}

func main() {
	artifactory := Artifactory{}
	artifactory.User = os.Getenv("ARTIFACTORY_USER")
	artifactory.Pass = os.Getenv("ARTIFACTORY_PASSWORD")
	artifactory.URL = os.Getenv("ARTIFACTORY_STORAGE_API")
	ArtifactoryBucket := os.Getenv("ARTIFACTORY_BUCKET")
	ArtifactoryBucketPath := os.Getenv("ARTIFACTORY_META_WRITE_PATH")
	// comma separated list of docker repositories
	// e.g. "dockerio,ghcr-io"
	ArtifactoryRepoList := os.Getenv("ARTIFACTORY_REPOLIST")
	if artifactory.URL == "" || artifactory.User == "" || ArtifactoryRepoList == "" {
		panic("ARTIFACTORY_STORAGE_API, ARITFACTORY_REPOLIST or ARTIFACTORY_USER cannot be empty")
	}
	repos := strings.Split(ArtifactoryRepoList, ",")
	var paths []string
	for _, r := range repos {
		p, err := IteratePath(artifactory, r)
		paths = append(paths, p...)
		check(err)
	}

	data := map[string]string{}
	for _, p := range paths {
		body, err := getData(p, artifactory.User, artifactory.Pass)
		check(err)
		fileInfo := File{}
		err = json.Unmarshal(body, &fileInfo)
		check(err)
		generateMeta(p, fileInfo.Checksums, data)
	}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Failed to marshal json")
		return
	}
	err = writeToS3(ArtifactoryBucket, ArtifactoryBucketPath, jsonStr)
	if err != nil {
		fmt.Printf("Failed to upload metadata to s3: %s\n", err)
	}
	return
}
