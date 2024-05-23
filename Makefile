.PHONY: test
test:
	go clean -testcache
	go test -race -v .

bench:
	go test -bench=.

m:
	rm -rf build/*.go
	mkdir -p build
	cgmerge
	mv _merged.go build/ritsu.go
	go fmt build/ritsu.go
	sed -i '' -e 's/package main/package ritsu/' build/ritsu.go

