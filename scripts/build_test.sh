#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Set the CGO flags to use the portable version of BLST
#
# We use "export" here instead of just setting a bash variable because we need
# to pass this flag to all child processes spawned by the shell.
export CGO_CFLAGS="-O -D__BLST_PORTABLE__"

go test -v -tags=test -race -timeout="3m" -coverprofile="./coverage.out" -covermode="atomic" ./...
