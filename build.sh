#!/bin/bash

set -euo pipefail
cd "$(dirname "$0")"

# Build artifacts
BUILD_ARTIFACT=m2cp BUILD_VERSION=5.17.1 BUILD_CORE_VERSION=24 BUILD_COMMIT_HASH=b882a80298fe4ffa8170e5e0f461ef56d51a8228 BUILD_VARIANT=arm64 BUILD_ARCH=arm64 src/build.sh

# Export artifacts
rm -rf ./bin
mkdir -p ./bin
mv ./src/bin/* ./bin/ 2>/dev/null || true

echo "✅ Artifacts exported to ./bin"
