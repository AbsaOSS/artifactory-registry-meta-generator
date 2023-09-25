package main

import (
	"testing"

	"github.com/AbsaOSS/artifactory-registry-meta-generator/artifactory"
	"github.com/stretchr/testify/assert"
)

func assertMaps[T, U comparable](m1 map[T]U, m2 map[T]U) bool {
	if len(m1) != len(m2) {
		return false
	}
	if m1 == nil && m2 != nil || m1 != nil && m2 == nil {
		return false
	}
	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}
	return true
}

func TestGenerateMainfestMeta(t *testing.T) {
	data := make(map[string]string)
	fInfo := artifactory.File{
		Path: "/thanos/thanos/v0.31.0/list.manifest.json",
		Checksums: artifactory.Checksum{
			SHA1:   "81a43c29d4f6cda7180fd4b59b7ce50ae6243f8e",
			MD5:    "084a2831a0ffa7eeac9a91e2a172cd26",
			SHA256: "e7d337d6ac2aea3f0f9314ec9830291789e16e2b480b9d353be02d05ce7f2a7e",
		},
	}
	expected := map[string]string{
		"/docker/registry/v2/blobs/sha256/e7/e7d337d6ac2aea3f0f9314ec9830291789e16e2b480b9d353be02d05ce7f2a7e/data":                                        "81a43c29d4f6cda7180fd4b59b7ce50ae6243f8e",
		"/docker/registry/v2/repositories/thanos/thanos/_manifests/revisions/sha256/e7d337d6ac2aea3f0f9314ec9830291789e16e2b480b9d353be02d05ce7f2a7e/link": "e7d337d6ac2aea3f0f9314ec9830291789e16e2b480b9d353be02d05ce7f2a7e",
		"/docker/registry/v2/repositories/thanos/thanos/_manifests/tags/v0.31.0/current/link":                                                              "e7d337d6ac2aea3f0f9314ec9830291789e16e2b480b9d353be02d05ce7f2a7e",
	}

	generateMeta(fInfo, data)
	assert.Len(t, data, 3)
	assert.True(t, assertMaps(data, expected))
}

func TestGenerateBlobMeta(t *testing.T) {
	data := make(map[string]string)
	fInfo := artifactory.File{
		Path: "/thanos/thanos/sha256__c02f71e18dcecb69d4ce396ddbbe53829330146996baa09a41602152aa55742b/sha256__05a2d9e5b341387ae9426a3040b6be2f33e5695a7ade88916f5990ca69b16522",
		Checksums: artifactory.Checksum{
			SHA1:   "6d3eae69ce0d84337d9c098c032a1c73476df552",
			MD5:    "b96120fc2997163478a48b81d20ce4eb",
			SHA256: "05a2d9e5b341387ae9426a3040b6be2f33e5695a7ade88916f5990ca69b16522",
		},
	}
	expected := map[string]string{
		"/docker/registry/v2/blobs/sha256/05/05a2d9e5b341387ae9426a3040b6be2f33e5695a7ade88916f5990ca69b16522/data":                           "6d3eae69ce0d84337d9c098c032a1c73476df552",
		"/docker/registry/v2/repositories/thanos/thanos/_layers/sha256/05a2d9e5b341387ae9426a3040b6be2f33e5695a7ade88916f5990ca69b16522/link": "05a2d9e5b341387ae9426a3040b6be2f33e5695a7ade88916f5990ca69b16522",
	}

	generateMeta(fInfo, data)
	assert.Len(t, data, 2)
	assert.True(t, assertMaps(data, expected))

}
