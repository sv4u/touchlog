.PHONY: default
BUILD_TIME := $(shell date +"%Y-%m-%d.%H:%M:%S")
BUILD_FLAG := "-X main.buildTime=${BUILD_TIME}"
GIT_VERSION := $(shell git describe --tags --abbrev=0)
PUBLISH_PATH = := "github.com/sv4u/touchlog@${GIT_VERSION}"

touchlog: main.go
	go build -v -ldflags=${BUILD_FLAG}

install:
	go install -v -ldflags=${BUILD_FLAG}
	touchlog --version --verbose

docs:
	[ -d dist ] || mkdir -p dist
	pandoc src/touchlog.1.md -s -t man -o dist/touchlog.1
	pandoc dist/touchlog.1 --from man --to html -s -o dist/touchlog.1.html
	pandoc README.md -s -t html -o dist/README.html

clean:
	-rm -rf dist
	-rm -rf touchlog

package: touchlog docs
	cp README.md dist
	cp LICENSE dist
	cp touchlog dist
	cp -r src dist
	cd dist && tar cvf touchlog-${GIT_VERSION}.tar .

publish: package
	GOPROXY=proxy.golang.org go list -m ${PUBLISH_PATH}
	./ftp.sh

default: build

