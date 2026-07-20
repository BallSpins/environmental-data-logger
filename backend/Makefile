.PHONY: test test-race test-cover run \
        stress-test bench-chunk bench-db bench-all

test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	go run ./cmd/app/main.go

stress-test:
	go run ./cmd/stress-shadow-device/main.go

bench-chunk:
	go test -bench=. ./cmd/benchmark-chunk/main.go

bench-db:
	go test -bench=. ./cmd/benchmark-db/main.go

bench-all:
	go run ./cmd/run-all/main.go

clean:
	go clean
	rm -rf results/*