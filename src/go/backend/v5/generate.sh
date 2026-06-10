#!/bin/bash
cd $(dirname $0)
pushd codegen
go run github.com/Khan/genqlient
popd
echo "Generated code in $(pwd)/generated.go"