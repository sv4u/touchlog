.PHONY: default

touchlog:
	gcc ,touchlog.c -o ,touchlog
	./,touchlog -v
	echo "OK"

optimized:
	gcc -O3 ,touchlog.c -o ,touchlog
	./,touchlog -v
	echo "OK"

documentation:
	pandoc ,touchlog.1.md -s -t man -o ,touchlog.1
	pandoc ,touchlog.1.md -s -t html -o ,touchlog.1.html
	pandoc README.md -s -t html -o README.html

clean:
	-rm ,touchlog
	-rm ,touchlog.1
	-rm ,touchlog.1.html
	-rm README.html

publish: optimized documentation

default: touchlog

all: touchlog documentation

