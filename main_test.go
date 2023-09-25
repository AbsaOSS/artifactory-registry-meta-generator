package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockArtifactoryApi(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpStatus := http.StatusOK
		path := "./testdata" + r.URL.Path
		body, err := os.ReadFile(path + ".json")
		if err != nil {
			httpStatus = http.StatusNotFound
		}
		w.WriteHeader(httpStatus)
		w.Write([]byte(body))
		w.Header().Set("Content-Type", "application/json")
	}
}

func TestUrlWalk(t *testing.T) {
	srv := httptest.NewServer(mockArtifactoryApi(t))
	arti := Artifactory{}
	arti.URL = srv.URL
	arti.User = "user"
	arti.Pass = "pass"
	defer srv.Close()
	paths, err := IteratePath(arti, "/quay-io")
	assert.NoError(t, err)
	assert.Len(t, paths, 7)
	for _, p := range paths {
		body, err := getData(p, arti.User, arti.Pass)
		assert.NoError(t, err)
		info := File{}
		err = json.Unmarshal(body, &info)
		assert.NoError(t, err)
		t.Logf("%s\n", p)

	}

}

func TestFileinfo(t *testing.T) {

}
