# Metadata generator for [docker distribution with artifactory backend](https://github.com/AbsaOSS/docker-distribution-artifactory)
Required environment variables:
* ARTIFACTORY_STORAGE_API - api path e.g. `/storage`
* ARTIFACTORY_USER - An Artifactory user to perform API calls with
* ARTIFACTORY_BUCKET - A S3 bucket used to store generated metadata
* ARTIFACTORY_META_WRITE_PATH - A path on S3 where to write metadata into
* ARTIFACTORY_REPOLIST - Artifactory docker repo list to scrape metadata for (e.g. "/quay.io")
