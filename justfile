bin_dir := env_var("HOME") / ".local/bin"

# list recipes
default:
    @just --list

# run all tests
test:
    go test ./...

# vet
vet:
    go vet ./...

# vet and test
check: vet test

# build every tool into dist/
build:
    mkdir -p dist
    go build -o dist/ ./cmd/...

# run a tool during development, e.g. `just run music-mirror --src /a`
run tool *args:
    go run ./cmd/{{tool}} {{args}}

# test, then build every tool into ~/.local/bin
install: test
    mkdir -p {{bin_dir}}
    go build -o {{bin_dir}}/ ./cmd/...

# pull latest
pull:
    git pull --ff-only

# on the server: pull, test, and, install
deploy: pull install