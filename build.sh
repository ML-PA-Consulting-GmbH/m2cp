#!/bin/bash

set -euo pipefail
cd "$(dirname "$0")"

# Build artifacts
BUILD_ARTIFACT=m2cp BUILD_VERSION=5.18.0 BUILD_CORE_VERSION=24 BUILD_COMMIT_HASH=dc607b5ada3bdc5fb49d0e7bc9c5caf88e3a51ef BUILD_VARIANT=arm64 BUILD_ARCH=arm64 src/build.sh

# Export artifacts
rm -rf ./bin
mkdir -p ./bin
mv ./src/bin/* ./bin/ 2>/dev/null || true

echo "✅ Artifacts exported to ./bin"
