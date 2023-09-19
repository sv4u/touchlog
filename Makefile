.PHONY: default

touchlog:
	gcc src/touchlog.c -o dist/,touchlog
	./dist/,touchlog -v
	echo "OK"

optimized:
	gcc -O3 src/,touchlog.c -o dist/,touchlog
	./dist/,touchlog -v
	echo "OK"

documentation:
	pandoc docs/touchlog.1.md -s -t man -o dist/,touchlog.1
	pandoc docs/touchlog.1.md -s -t html -o dist/,touchlog.1.html

clean:
	-rm -rf dist

publish: optimized documentation

default: touchlog

all: touchlog documentation

