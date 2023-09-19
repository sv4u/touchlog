.PHONY: default

touchlog:
	[ -d dist ] || mkdir -p dist
	gcc src/touchlog.c -o dist/,touchlog
	./dist/,touchlog -v
	echo "OK"

optimized:
	[ -d dist ] || mkdir -p dist
	gcc -O3 src/touchlog.c -o dist/,touchlog
	./dist/,touchlog -v
	echo "OK"

documentation:
	[ -d dist ] || mkdir -p dist
	pandoc docs/touchlog.1.md -s -t man -o dist/,touchlog.1
	pandoc docs/touchlog.1.md -s -t html -o dist/,touchlog.1.html

clean:
	-rm -rf dist

publish: optimized documentation

default: touchlog

all: touchlog documentation

