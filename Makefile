.PHONY: default
BUILD_TIME := $(shell date +"%Y-%m-%d.%H:%M:%S")
GIT_VERSION := $(shell git describe --tags --abbrev=0)
BUILD_FLAG := "-X main.buildTime=${BUILD_TIME} -X main.version=${GIT_VERSION}"
GH_PUBLISH_PATH := "github.com/sv4u/touchlog@${GIT_VERSION}"

HOST := "cpanel.freehosting.com"
UNAME := "sasankvi"
PASSWD := $(shell echo ${WEBSITE_ENC_KEY} | base64 --decode)
WEB_PATH := "domains/development.sasankvishnubhatla.net/public_html/log-suite/touchlog/"

touchlog: touchlog.go
	go build -v -ldflags=${BUILD_FLAG}

install:
	go install -v -ldflags=${BUILD_FLAG}

docs:
	[ -d dist ] || mkdir -p dist
	pandoc touchlog.md -s -t man -o dist/touchlog.1
	pandoc dist/touchlog.1 --from man --to html -s -o dist/touchlog.1.html
	pandoc README.md -s -t html -o dist/README.html

clean:
	-rm -rf dist
	-rm -rf touchlog

package: touchlog docs
	cp README.md dist
	cp LICENSE dist
	cp touchlog dist
	cp touchlog.go dist
	cp go.mod dist

publish: package
	GOPROXY=proxy.golang.org go list -m ${GH_PUBLISH_PATH}

default: touchlog

