.PHONY: default
BUILD_TIME := $(shell date +"%Y-%m-%d.%H:%M:%S")
FLAG := "-X main.buildTime=${BUILD_TIME}"

build:
	echo $(BUILD_TIME)
	echo $(FLAG)
	go build -v -ldflags=${FLAG}
	echo "OK"

documentation:
	[ -d dist ] || mkdir -p dist
	pandoc docs/touchlog.1.md -s -t man -o dist/,touchlog.1
	pandoc dist/,touchlog.1 --from man --to html -s -o dist/,touchlog.1.html
	pandoc README.md -s -t html -o dist/README.html
	echo "OK"

clean:
	-rm -rf dist && echo "Clean"

publish: build documentation
	cp -r src dist
	cp README.md dist
	cp LICENSE dist
	echo "OK"

default: build

