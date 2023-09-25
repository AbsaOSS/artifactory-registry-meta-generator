package artifactory

import (
	"fmt"
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

func TestGetFilePaths(t *testing.T) {
	const expected = `[{Checksums:{SHA1:81a43c29d4f6cda7180fd4b59b7ce50ae6243f8e MD5:084a2831a0ffa7eeac9a91e2a172cd26 SHA256:e7d337d6ac2aea3f0f9314ec9830291789e16e2b480b9d353be02d05ce7f2a7e} Path:/thanos/thanos/sha256__e7d337d6ac2aea3f0f9314ec9830291789e16e2b480b9d353be02d05ce7f2a7e/list.manifest.json} {Checksums:{SHA1:075760c4412592f8b8ff2482dbc2dd626e609677 MD5:55d96b3865f3ab2e47613e4e25af6d7b SHA256:c02f71e18dcecb69d4ce396ddbbe53829330146996baa09a41602152aa55742b} Path:/thanos/thanos/sha256__c02f71e18dcecb69d4ce396ddbbe53829330146996baa09a41602152aa55742b/manifest.json} {Checksums:{SHA1:6d3eae69ce0d84337d9c098c032a1c73476df552 MD5:b96120fc2997163478a48b81d20ce4eb SHA256:05a2d9e5b341387ae9426a3040b6be2f33e5695a7ade88916f5990ca69b16522} Path:/thanos/thanos/sha256__c02f71e18dcecb69d4ce396ddbbe53829330146996baa09a41602152aa55742b/sha256__05a2d9e5b341387ae9426a3040b6be2f33e5695a7ade88916f5990ca69b16522} {Checksums:{SHA1:2530d086e6ef0d289d21ccb70b5e79676cd5a75d MD5:39a0c9f2edf0cc9ffc868c81ae4e8d78 SHA256:765df1804bae188b4f5d2326283768dd98a908883f0b8e9c85e192f294ea2309} Path:/thanos/thanos/sha256__c02f71e18dcecb69d4ce396ddbbe53829330146996baa09a41602152aa55742b/sha256__765df1804bae188b4f5d2326283768dd98a908883f0b8e9c85e192f294ea2309} {Checksums:{SHA1:886f392c130d13dd5eec69b73cea92a87360c680 MD5:e25bd5d4467ae9e73b5bef257a543c1e SHA256:9b231b23b5cdc7c22cc3c519df0e35876200ecd02e69978ab7dfc2fad43b384e} Path:/thanos/thanos/sha256__c02f71e18dcecb69d4ce396ddbbe53829330146996baa09a41602152aa55742b/sha256__9b231b23b5cdc7c22cc3c519df0e35876200ecd02e69978ab7dfc2fad43b384e} {Checksums:{SHA1:bcf48edb6c4d066a071374e4e7255baaac4375f9 MD5:2d7b372540ff6e6d24971305c5b27248 SHA256:a1e445c9ea057d36599d14854b3d9b1be3087dd76916eab071f221c15147d66f} Path:/thanos/thanos/sha256__c02f71e18dcecb69d4ce396ddbbe53829330146996baa09a41602152aa55742b/sha256__a1e445c9ea057d36599d14854b3d9b1be3087dd76916eab071f221c15147d66f} {Checksums:{SHA1:81a43c29d4f6cda7180fd4b59b7ce50ae6243f8e MD5:084a2831a0ffa7eeac9a91e2a172cd26 SHA256:e7d337d6ac2aea3f0f9314ec9830291789e16e2b480b9d353be02d05ce7f2a7e} Path:/thanos/thanos/v0.31.0/list.manifest.json}]`
	srv := httptest.NewServer(mockArtifactoryApi(t))
	arti := Artifactory{}
	arti.URL = srv.URL
	arti.User = "user"
	arti.Pass = "pass"
	arti.RepoList = []string{"/quay-io"}
	defer srv.Close()
	fileInfo, _ := arti.GetFilePaths()
	assert.Equal(t, expected, fmt.Sprintf("%+v", fileInfo))
}
