#!/bin/bash

set -euo pipefail
cd "$(dirname "$0")"

# Build artifacts
BUILD_ARTIFACT=m2cp BUILD_VERSION=5.19.8 BUILD_CORE_VERSION=24 BUILD_COMMIT_HASH=f19fbfdf8451c27b4704d6ca283777539468a7a9 BUILD_VARIANT=arm64 BUILD_ARCH=arm64 src/build.sh

# Export artifacts
rm -rf ./bin
mkdir -p ./bin
mv ./src/bin/* ./bin/ 2>/dev/null || true

echo "✅ Artifacts exported to ./bin"
