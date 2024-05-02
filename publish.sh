#!/bin/bash
echo touchlog golang publisher

echo getting latest annotated git tag as version to publish
tag=$(git describe --tags --abbrev=0)
echo "tag: ${tag}"

repo="gitlab.com/log-suite/touchlog"
golang_repo=${repo}"@"${tag}

echo "golang repo: ${golang_repo}"

GOPROXY=proxy.golang.org go list -m ${golang_repo}
