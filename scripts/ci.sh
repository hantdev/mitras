#!/bin/bash
set -e

NPROC=$(sysctl -n hw.logicalcpu)
GO_VERSION=1.24.2
PROTOC_VERSION=30.2
PROTOC_GEN_VERSION=v1.36.6
PROTOC_GRPC_VERSION=v1.5.1
GOLANGCI_LINT_VERSION=v2.0.2

function version_gt() { test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"; }

update_go() {
    echo "Checking Go version..."
    if command -v go >/dev/null 2>&1; then
        CURRENT_GO_VERSION=$(go version | sed 's/[^0-9.]*\([0-9.]*\).*/\1/')
    else
        CURRENT_GO_VERSION="0"
    fi

    if version_gt $GO_VERSION $CURRENT_GO_VERSION; then
        echo "Updating Go from $CURRENT_GO_VERSION to $GO_VERSION..."
        rm -rf /usr/local/go $HOME/go go$GO_VERSION.darwin-arm64.tar.gz || true
        curl -LO https://go.dev/dl/go$GO_VERSION.darwin-arm64.tar.gz
        sudo tar -C /usr/local -xzf go$GO_VERSION.darwin-arm64.tar.gz
        rm go$GO_VERSION.darwin-arm64.tar.gz
    fi
    export GOROOT=/usr/local/go
    export GOPATH=$HOME/go
    export GOBIN=$GOPATH/bin
    export PATH=$GOROOT/bin:$GOBIN:$PATH
    go version
}

setup_protoc() {
    echo "Setting up protoc..."
    PROTOC_ZIP=protoc-$PROTOC_VERSION-osx-aarch_64.zip
    curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/$PROTOC_ZIP
    unzip -o $PROTOC_ZIP -d protoc3

    sudo cp -f protoc3/bin/* /usr/local/bin/

    # Clean & Copy protobuf include only
    sudo mkdir -p /usr/local/include/google
    sudo rm -rf /usr/local/include/google/protobuf
    sudo cp -R protoc3/include/google/protobuf /usr/local/include/google/

    rm -rf $PROTOC_ZIP protoc3

    go install google.golang.org/protobuf/cmd/protoc-gen-go@$PROTOC_GEN_VERSION
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$PROTOC_GRPC_VERSION
}

setup_lint() {
    echo "Setting up golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/main/install.sh | sh -s -- -b $(go env GOBIN) $GOLANGCI_LINT_VERSION
}

setup_mg() {
    echo "Checking proto Mitras..."
    for p in $(find . -name "*.pb.go"); do mv "$p" "$p.tmp"; done
    make proto
    for p in $(find . -name "*.pb.go"); do
        if ! cmp -s "$p" "$p.tmp"; then
            echo "Mismatch in generated file: $p"
            exit 1
        fi
    done
    echo "Compile check for rabbitmq..."
    MITRAS_MESSAGE_BROKER_TYPE=rabbitmq make http
    echo "Compile check for redis..."
    MITRAS_ES_TYPE=redis make http
    make -j$NPROC
}

run_test() {
    # echo "Running lint..."
    # golangci-lint run
    echo "Running tests..."
    echo "" > coverage.txt
    for d in $(go list ./... | grep -v 'vendor\|cmd'); do
        GOCACHE=off go test -mod=vendor -v -race -tags test -coverprofile=profile.out -covermode=atomic $d
        if [ -f profile.out ]; then
            cat profile.out >> coverage.txt
            rm profile.out
        fi
    done
}

push() {
    if [[ "$BRANCH_NAME" == "main" ]]; then
        echo "Pushing Docker images..."
        make -j$NPROC latest
    fi
}

setup() {
    echo "Setting up environment..."
    # update_go
    # setup_protoc
    setup_mg
    setup_lint
}

setup
run_test
push
