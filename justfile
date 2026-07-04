default:
    @just --list

# build the respec binary into build/
build:
    go build -o build/respec .

# build and install the binary into ~/.local/bin/
install: build
    install -D build/respec ~/.local/bin/respec

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

# bump the version, commit, and tag (patch|minor|major|...)
bump part="patch":
    go tool versionbump {{part}}
