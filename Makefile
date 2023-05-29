TEST_PKG?=./...

lint-prepare:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s latest

lint:
	test -f ./bin/golangci-lint || make lint-prepare
	./bin/golangci-lint run -v --timeout=1m ./...

fmt:
	@go fmt ./...

test:
	@go test -v -race -timeout=10m ${TEST_PKG} 

bench:
	@go test -benchmem -bench ^Benchmark* ./benchmark/...