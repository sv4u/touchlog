.PHONY: default
BUILD_TIME := $(shell date +"%Y-%m-%d.%H:%M:%S")
FLAG := "-X main.buildTime=${BUILD_TIME}"

build:
	[ -d dist ] || mkdir -p dist
	go build -v -ldflags=${FLAG}
	cp touchlog dist
	./touchlog --version --verbose

docs:
	[ -d dist ] || mkdir -p dist
	pandoc src/touchlog.1.md -s -t man -o dist/touchlog.1
	pandoc dist/touchlog.1 --from man --to html -s -o dist/touchlog.1.html
	pandoc README.md -s -t html -o dist/README.html

clean:
	-rm -rf dist

publish: build docs
	cp -r src dist
	cp README.md dist
	cp LICENSE dist

default: build

