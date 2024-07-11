.PHONY: default
BUILD_TIME := $(shell date +"%Y-%m-%d.%H:%M:%S")
GIT_VERSION := $(shell git describe --tags --abbrev=0)
GIT_HASH := $(shell git rev-parse --short=8 @)

BUILD_FLAG := "-X main.buildTime=${BUILD_TIME} -X main.version=${GIT_VERSION}"
GH_PUBLISH_PATH := "github.com/sv4u/touchlog@${GIT_VERSION}"

HOST := "cpanel.freehosting.com"
UNAME := "sasankvi"
WEB_PATH := "domains/development.sasankvishnubhatla.net/public_html/log-suite/"

touchlog: touchlog.go
	go build -v -ldflags=${BUILD_FLAG}

install: docs
	go install -v -ldflags=${BUILD_FLAG}
	mkdir -p /usr/local/share/man/man1
	cp dist/touchlog.1 /usr/local/share/man/man1/
	mandb

docs:
	[ -d dist ] || mkdir -p dist
	pandoc manpage.md -s -t man -o dist/touchlog.1
	pandoc dist/touchlog.1 --from man --to html -s -o dist/touchlog.1.html
	pandoc README.md -s -t html -o dist/README.html
	pandoc logplate.md -s -t html -o dist/logplate.html

clean:
	-rm -rf dist
	-rm -rf touchlog

package: touchlog docs
	cp README.md dist
	cp logplate.md dist
	cp LICENSE dist
	mv touchlog dist
	cp touchlog.go dist
	cp go.mod dist
	cp -r templates dist

publish: package
	GOPROXY=proxy.golang.org go list -m ${GH_PUBLISH_PATH}

dtarballs: package
	tar cvf dist/touchlog-${GIT_HASH}-bin.tar -C dist README.md touchlog LICENSE templates
	tar cvf dist/touchlog-${GIT_HASH}-src.tar -C dist README.md touchlog LICENSE touchlog.1 touchlog.go templates

ptarballs: package
	tar cvf dist/touchlog-${GIT_VERSION}-bin.tar -C dist README.md touchlog LICENSE templates
	tar cvf dist/touchlog-${GIT_VERSION}-src.tar -C dist README.md touchlog LICENSE touchlog.1 touchlog.go templates

website: ptarballs
	mv dist touchlog
	$(eval PASSWD := $(shell echo ${WEBSITE_ENC_KEY} | base64 --decode))
	ncftpput -R -u ${UNAME} -p ${PASSWD} ${HOST} ${WEB_PATH} touchlog/
	mv touchlog dist

default: touchlog
