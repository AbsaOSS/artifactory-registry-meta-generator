package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AbsaOSS/artifactory-registry-meta-generator/artifactory"
)

const (
	dockerBaseURL = "/docker/registry/v2"
)

// /bks-docker-local/stefanprodan/podinfo/5.2.0/manifest.json ->
// /bks-docker-local/stefanprodan/podinfo/_manifests/tags/5.2.0/current/link
func normalizePath(p string) string {
	return strings.TrimPrefix(p, os.Getenv("ARTIFACTORY_URL"))
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
func getLinkToBlob(sum artifactory.Checksum, data map[string]string) {
	shaPath := sha256path(sum.SHA256)
	data[dockerBaseURL+fmt.Sprintf("/blobs/sha256/%s/data", shaPath)] = sum.SHA1
}

func sha256path(s string) string {
	return fmt.Sprintf("%s/%s", s[0:2], s)
}

// /bks-docker-local/cert-manager-controller/v0.12.0-venafi/sha256__c5c9eab06e7db9e76641b4fe8351725d5d3d40100db3f0efaa411807022441e6 ->
// /repositories/bks-docker-local/cert-manager-controller/_layers/sha256/b02a7525f878e61fc1ef8a7405a2cc17f866e8de222c1c98fd6681aff6e509db/link
func handleLayerLink(repo, s string, sha256 string) string {
	regex := regexp.MustCompile("(/sha256__.*)")
	image := regex.ReplaceAllString(s, "")
	return dockerBaseURL + fmt.Sprintf("/repositories/%s%s/_layers/sha256/%s/link", repo, image, sha256)
}

func generateManifest(s string, info artifactory.File, data map[string]string) {
	sum := info.Checksums
	pathParts := strings.Split(s, "/")
	partsLen := len(pathParts)
	tag := pathParts[partsLen-2]
	image := strings.Join(pathParts[0:partsLen-2], "/")
	// link withou sha256
	link := fmt.Sprintf("/repositories/%s%s/_manifests/tags/%s/current/link", info.Repo, image, tag)
	revLink := convertLinkToRev(link, sum.SHA256)
	data[dockerBaseURL+link] = sum.SHA256
	data[revLink] = sum.SHA256
	getLinkToBlob(info.Checksums, data)
}
func generateLayer(s string, info artifactory.File, data map[string]string) {
	sha256 := info.Checksums.SHA256
	data[handleLayerLink(info.Repo, s, sha256)] = sha256
	getLinkToBlob(info.Checksums, data)
}
func generateMeta(f artifactory.File, data map[string]string) {
	f.Repo = strings.TrimSuffix(f.Repo, "-cache")
	p := normalizePath(f.Path)
	if isManifest(p) {
		generateManifest(p, f, data)
	}
	if isLayer(p) {
		generateLayer(p, f, data)
	}
}
