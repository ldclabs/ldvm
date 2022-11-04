#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CGO_ENABLED=1 go test -v -race -timeout="3m" -coverprofile="./coverage.out" -covermode="atomic" ./...
