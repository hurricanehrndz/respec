default:
    @just --list

# build the respec binary into build/
build:
    go build -o build/respec .

# vet + run all tests
test:
    go vet ./...
    go test ./...

# run golangci-lint over the module
lint:
    golangci-lint run ./...

# remove build artifacts
clean:
    rm -rf build
