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
func getLinkToBlob(file artifactory.Checksum, data map[string]string) {
	shaPath := sha256path(file.SHA256)
	data[dockerBaseURL+fmt.Sprintf("/blobs/sha256/%s/data", shaPath)] = file.SHA1
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

func generateManifest(s string, info artifactory.Checksum, data map[string]string) {
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
func generateLayer(s string, info artifactory.Checksum, data map[string]string) {
	data[handleLayerLink(s, info.SHA256)] = info.SHA256
	getLinkToBlob(info, data)
}
func generateMeta(f artifactory.File, data map[string]string) {
	p := normalizePath(f.Path)
	if isManifest(p) {
		generateManifest(p, f.Checksums, data)
	}
	if isLayer(p) {
		generateLayer(p, f.Checksums, data)
	}
}
