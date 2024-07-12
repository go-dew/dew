.PHONY: test
test:
	go clean -testcache
	go test -race -v .

bench:
	go test -bench=.
