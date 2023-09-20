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
	pandoc dist/,touchlog.1 --from man --to html -s -o dist/,touchlog.1.html
	pandoc README.md -s -t html -o dist/README.html
	echo "OK"

clean:
	-rm -rf dist && echo "Clean"

publish: optimized documentation
	cp -r src dist
	cp README.md dist
	cp LICENSE dist
	echo "OK"

default: touchlog

