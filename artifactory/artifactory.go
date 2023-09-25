package artifactory

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Artifactory struct {
	URL      string
	User     string
	Pass     string
	RepoList []string
}

type paths struct {
	Paths []path `json:"children"`
}

type path struct {
	Folder bool   `json:"folder"`
	Uri    string `json:"uri"`
}

type File struct {
	Checksums Checksum `json:"checksums"`
	Path      string   `json:"path"`
}

type Checksum struct {
	SHA1   string `json:"sha1"`
	MD5    string `json:"md5"`
	SHA256 string `json:"sha256"`
}

func CreateArtifactory() (*Artifactory, error) {
	artifactory := &Artifactory{}
	artifactory.User = os.Getenv("ARTIFACTORY_USER")
	artifactory.Pass = os.Getenv("ARTIFACTORY_PASSWORD")
	artifactory.URL = os.Getenv("ARTIFACTORY_STORAGE_API")
	// comma separated list of docker repositories
	// e.g. "/dockerio,/ghcr-io"
	artifactory.RepoList = strings.Split(os.Getenv("ARTIFACTORY_REPOLIST"), ",")
	if artifactory.URL == "" || artifactory.User == "" || len(artifactory.RepoList) == 0 {
		return nil, fmt.Errorf("ARTIFACTORY_STORAGE_API, ARITFACTORY_REPOLIST or ARTIFACTORY_USER cannot be empty")
	}
	return artifactory, nil
}

func (a *Artifactory) GetFilePaths() ([]File, error) {
	var paths []string
	var files []File
	for _, r := range a.RepoList {
		p, err := a.iteratePath(r)
		paths = append(paths, p...)
		if err != nil {
			return nil, err
		}
	}

	for _, p := range paths {
		fileInfo := File{}
		body, err := a.getData(p)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &fileInfo)
		if err != nil {
			return nil, err
		}
		files = append(files, fileInfo)
	}
	return files, nil
}

func (a *Artifactory) iteratePath(p string) ([]string, error) {
	body, err := a.getData(a.URL + p)
	if err != nil {
		return nil, err
	}
	list := paths{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, path := range list.Paths {
		if !path.Folder {
			paths = append(paths, []string{a.URL + p + path.Uri}...)
			continue
		}

		p, err := a.iteratePath(p + path.Uri)
		if err != nil {
			return nil, err
		}
		paths = append(paths, p...)
	}
	return paths, nil
}

func (a *Artifactory) getData(uri string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(a.User, a.Pass)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
