.PHONY: test
test:
	@go clean -testcache
	@go test -race -v -coverprofile="coverage.txt" -covermode=atomic ./...

.PHONY: test-coverage
open-coverage:
	@go tool cover -html=coverage.txt

bench:
	@go test -bench=.

