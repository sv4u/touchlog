.PHONY: default
BUILD_TIME := $(shell date +"%Y-%m-%d.%H:%M:%S")
GIT_VERSION := $(shell git describe --tags --abbrev=0)
GIT_HASH := $(shell git rev-parse --short=8 @)

BUILD_FLAG := "-X main.buildTime=${BUILD_TIME} -X main.version=${GIT_VERSION}"
GH_PUBLISH_PATH := "github.com/sv4u/touchlog@${GIT_VERSION}"

HOST := "cpanel.freehosting.com"
UNAME := "sasankvi"
PASSWD := $(shell echo ${WEBSITE_ENC_KEY} | base64 --decode)
WEB_PATH := "domains/development.sasankvishnubhatla.net/public_html/log-suite/touchlog/"

touchlog: touchlog.go
	go build -v -ldflags=${BUILD_FLAG}

install: docs
	go install -v -ldflags=${BUILD_FLAG}
	mkdir -p /usr/local/share/man/man1
	cp dist/touchlog.1 /usr/local/share/man/man1/
	mandb

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

dtarballs: package
	tar cvf dist/touchlog-${GIT_HASH}-bin.tar -C dist README.md touchlog LICENSE
	tar cvf dist/touchlog-${GIT_HASH}-src.tar -C dist README.md touchlog LICENSE touchlog.1 touchlog.go

ptarballs: package
	tar cvf dist/touchlog-${GIT_VERSION}-bin.tar -C dist README.md touchlog LICENSE
	tar cvf dist/touchlog-${GIT_VERSION}-src.tar -C dist README.md touchlog LICENSE touchlog.1 touchlog.go

website: ptarballs
	cd dist && ncftpput -R -u ${UNAME} -p ${PASSWD} ${HOST} ${WEB_PATH} .

default: touchlog
