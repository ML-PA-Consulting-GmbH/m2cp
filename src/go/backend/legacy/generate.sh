#!/usr/bin/env bash

set -euxo pipefail

cd $(dirname $0)
pushd codegen
GOTOOLCHAIN=go1.24.10 go run github.com/Khan/genqlient@latest
popd
echo "Generated code in $(pwd)/generated.go"