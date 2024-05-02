#!/bin/bash
echo touchlog golang publisher

echo getting latest annotated git tag as version to publish
tag=$(git describe --tags --abbrev=0)
echo "tag: ${tag}"

repo="github.com/sv4u/touchlog@"
repo+="${tag}"

echo "golang repo: ${repo}"

GOPROXY=proxy.golang.org go list -m "${repo}"
