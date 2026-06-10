#!/usr/bin/bash
# m2cp build entry — librelease's default export() bundles this as the
# release's src/build.sh, then the generated bundle wrapper invokes it ONCE
# (a release isn't arch-specific). Looping over the supported archs is this
# script's responsibility; per-arch logic lives in go/build.sh.
set -euo pipefail
cd "$(dirname "$0")"

# Disable workspace mode so Go doesn't walk up into a parent go.work that
# doesn't list this bundle's module (e.g. when verifying locally inside the
# release-time repo). Customer extractions don't have a workspace either,
# so this is harmless there.
export GOWORK=off

for arch in arm64 amd64 win; do
  echo "==> Building ${BUILD_ARTIFACT:-m2cp} ${BUILD_VERSION:-?} ($arch)"
  BUILD_ARCH="$arch" ./go/build.sh
done

# Consolidate per-arch outputs under ./bin/ so the librelease wrapper's
# `mv src/bin/* bin/` and build_release's bin/ collection pick them up.
mkdir -p bin
[[ -d go/bin ]] && cp -a go/bin/. bin/
for d in go/build/linux-*/; do
  [[ -f "${d}m2cp" ]] || continue
  arch="${d%/}"; arch="${arch##*/linux-}"
  mkdir -p "bin/$arch"
  cp -a "${d}m2cp" "bin/$arch/m2cp"
done
[[ -d go/build/win-amd64 ]] && { mkdir -p bin/win-amd64 && cp -a go/build/win-amd64/. bin/win-amd64/; }
