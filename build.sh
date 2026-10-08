#!/bin/bash

set -euo pipefail
cd "$(dirname "$0")"

# Build artifacts
BUILD_ARTIFACT=m2cp BUILD_VERSION=5.18.4 BUILD_CORE_VERSION=24 BUILD_COMMIT_HASH=ecf50d9321870552417577e158c65ff60c6ff8b9 BUILD_VARIANT=arm64 BUILD_ARCH=arm64 src/build.sh

# Export artifacts
rm -rf ./bin
mkdir -p ./bin
mv ./src/bin/* ./bin/ 2>/dev/null || true

echo "✅ Artifacts exported to ./bin"
