#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

go test -v -race -timeout="3m" -coverprofile="./coverage.out" -covermode="atomic" ./...
