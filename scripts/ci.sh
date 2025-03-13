#!/bin/bash
# This script contains commands to be executed by the CI tool on macOS.

NPROC=$(command -v nproc >/dev/null && nproc || sysctl -n hw.ncpu)
GO_VERSION=1.22.4
PROTOC_VERSION=27.1
PROTOC_GEN_VERSION=v1.34.2
PROTOC_GRPC_VERSION=v1.4.0
GOLANGCI_LINT_VERSION=v1.60.3

function version_gt() { test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"; }

update_go() {
    CURRENT_GO_VERSION=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')
    if version_gt $GO_VERSION $CURRENT_GO_VERSION; then
        echo "Updating Go from $CURRENT_GO_VERSION to $GO_VERSION ..."
        sudo rm -rf /usr/local/go
        ARCH=$(uname -m | sed 's/x86_64/amd64/;s/arm64/arm64/')
        curl -OL https://go.dev/dl/go$GO_VERSION.darwin-$ARCH.tar.gz
        sudo tar -C /usr/local -xzf go$GO_VERSION.darwin-$ARCH.tar.gz
        rm go$GO_VERSION.darwin-$ARCH.tar.gz
        export PATH=/usr/local/go/bin:$PATH
    fi
    export GOBIN=$HOME/go/bin
    export PATH=$GOBIN:$PATH
    go version
}

setup_protoc() {
    echo "Setting up protoc..."
    PROTOC_ZIP=protoc-$PROTOC_VERSION-osx-x86_64.zip
    curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/$PROTOC_ZIP
    unzip -o $PROTOC_ZIP -d protoc3
    sudo mv protoc3/bin/* /usr/local/bin/
    sudo mv protoc3/include/* /usr/local/include/
    rm -rf $PROTOC_ZIP protoc3

    go install google.golang.org/protobuf/cmd/protoc-gen-go@$PROTOC_GEN_VERSION
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$PROTOC_GRPC_VERSION
}

setup_lint() {
    echo "Setting up GolangCI-Lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@$GOLANGCI_LINT_VERSION
}

setup() {
    echo "Setting up environment..."
    update_go
    setup_protoc
    setup_lint
}

run_test() {
    echo "Running lint..."
    golangci-lint run
    echo "Running tests..."
    truncate -s 0 coverage.txt
    for d in $(go list ./... | grep -v 'vendor\|cmd'); do
        GOCACHE=off
        go test -v -race -coverprofile=profile.out -covermode=atomic $d
        if [ -f profile.out ]; then
            cat profile.out >> coverage.txt
            rm profile.out
        fi
    done
}

push() {
    if [[ "${BRANCH_NAME:-}" == "main" ]]; then
        echo "Pushing Docker images..."
        make -j$NPROC latest
    fi
}

set -e
setup
run_test
push
