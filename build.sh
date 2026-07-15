#!/bin/bash

set -euo pipefail
cd "$(dirname "$0")"

# Build artifacts
BUILD_ARTIFACT=m2cp BUILD_VERSION=5.14.1 BUILD_CORE_VERSION=24 BUILD_COMMIT_HASH=3ea0e214ed9c7695b8f506330072631fd12b54ea BUILD_VARIANT=arm64 BUILD_ARCH=arm64 src/build.sh

# Export artifacts
rm -rf ./bin
mkdir -p ./bin
mv ./src/bin/* ./bin/ 2>/dev/null || true

echo "✅ Artifacts exported to ./bin"
