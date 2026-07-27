#!/usr/bin/bash
# NOTICE: when editing this build script in Windows, it will mess up line breaks

set -e          # exit on error
set -u          # exit on undefined variables
set -o pipefail # exit on command pipe failures

cd "$(dirname "$0")"

# Ignore any go.work upstream: this module is intentionally not a workspace
# member, and its go.mod already pins the m2cp SDK via a `replace` directive.
export GOWORK=off

# --- fallbacks for direct invocation (CI/release sets these explicitly) ---
host_arch=$(uname -m)
case "$host_arch" in
  x86_64) host_arch=amd64 ;;
  aarch64) host_arch=arm64 ;;
esac
: "${BUILD_ARCH:=$host_arch}"
: "${BUILD_ARTIFACT:=m2cp}"
: "${BUILD_VERSION:=0.0.0-dev}"

# This is the name you can use with e.g. `apt install`
PACKAGE_NAME=mlpa-m2cp-cli
DIR_BUILD_DEB=$(pwd)/build/deb/${BUILD_ARCH}
DIR_BIN=$(pwd)/bin/${BUILD_ARCH}

clean() {
  if [[ "$BUILD_ARCH" == "win" ]]; then
    rm -rf build/win-amd64/
  else
    rm -rf build/linux-${BUILD_ARCH}/
    rm -rf ${DIR_BUILD_DEB}
    rm -rf ${DIR_BIN}
  fi
}

prepare() {
  mkdir -p ${DIR_BUILD_DEB}
}

build() {
  # Prefer libbuild's docker-wrapped Go when run via librelease (deterministic,
  # no host go needed). Fall back to host `go` for standalone customer builds.
  # Either can be overridden by setting BUILD_GO explicitly.
  : "${BUILD_GO:=${DIR_LIBBUILD:+$DIR_LIBBUILD/go-u20}}"
  : "${BUILD_GO:=go}"

  # Commit hash is captured at release time (BUILD_COMMIT_HASH) — no git
  # available inside the bundled release tree. For direct invocation from a
  # checkout, fall back to `git rev-parse` so dev builds carry a real commit.
  local commit_hash="${BUILD_COMMIT_HASH:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"

  # go-uXX wrappers rebuild the command via `$*` and run it through bash -c,
  # which re-tokenises on spaces — so the ldflags value needs explicit inner
  # quoting to survive. Plain `go` must NOT see the inner quotes.
  # (Same dual quoting libbuild's compile() applies.)
  local ldflags="-s -w -X m2cpcli/version.Version=${BUILD_VERSION} -X m2cpcli/version.CommitHash=${commit_hash}"
  case "$(basename "$BUILD_GO")" in
    go-u*) ldflags="\"$ldflags\"" ;;
  esac

  if [[ "$BUILD_ARCH" == "win" ]]; then
    GOOS=windows GOARCH=amd64 "$BUILD_GO" build \
      -ldflags="$ldflags" \
      -o build/win-amd64/${BUILD_ARTIFACT}.exe
    windows/build.sh
    cp -f windows/build/m2cp-setup.exe build/win-amd64/
    echo "Created windows installer in build/win-amd64/m2cp-setup.exe"
  else
    # Note, this requires <module path (as in go.mod)>/<package name>.<VariableName>
    GOOS=linux GOARCH=$BUILD_ARCH "$BUILD_GO" build \
      -ldflags="$ldflags" \
      -o build/linux-${BUILD_ARCH}/${BUILD_ARTIFACT}
  fi
}

create_control_file() {
  touch control
  echo "Package: ${PACKAGE_NAME}" >>control
  echo "Version: ${BUILD_VERSION}" >>control
  echo "Section: dev" >>control
  echo "Priority: optional" >>control
  echo "Architecture: ${BUILD_ARCH}" >>control
  echo "Maintainer: M2CP Team <m2cp-support@ml-pa.com>" >>control
  echo "Description: command line interface" >>control
}

compose_debian_package() {
  # Compose the Debian package
  mkdir -p ${DIR_BUILD_DEB}
  pushd ${DIR_BUILD_DEB}

  DEBIAN_DIR="DEBIAN"
  PACKAGE_BUILD_DIR="${PACKAGE_NAME}_${BUILD_VERSION}_${BUILD_ARCH}"
  mkdir -p ${PACKAGE_BUILD_DIR}/${DEBIAN_DIR}

  pushd ${PACKAGE_BUILD_DIR}/${DEBIAN_DIR}
  create_control_file
  popd

  mkdir -p ${PACKAGE_BUILD_DIR}/usr/bin/
  cp ../../linux-${BUILD_ARCH}/m2cp "${PACKAGE_BUILD_DIR}/usr/bin/m2cp"
  chmod +x "${PACKAGE_BUILD_DIR}/usr/bin/m2cp"
  ln -sf m2cp "${PACKAGE_BUILD_DIR}/usr/bin/liot"

  # Note, `aptly` does not support zst compression
  dpkg-deb -Zxz --build ${PACKAGE_BUILD_DIR}

  popd

  mkdir -p ${DIR_BIN}
  mv -v ${DIR_BUILD_DEB}/${PACKAGE_BUILD_DIR}.deb ${DIR_BIN}
}

clean
prepare
build

if [[ "$BUILD_ARCH" != "win" ]]; then
  compose_debian_package
fi

# windows version doesn't compile because of snapd dependencies
#GOOS=windows GOARCH=amd64 m2cp-go build -o bin/amd64/m2cp.exe main.go

# -------------------------
# build tests  # TODO: why was that done like this prior 2023-10-16?
# -------------------------

#if [ "$build_tests" == "true" ]; then
#
#  TEST_BIN_DIR=bin/amd64/tests/
#
#  # Clear the directory with the test binaries
#  rm -rf $TEST_BIN_DIR
#  mkdir -p $TEST_BIN_DIR
#
#  # Compile all test binaries
#  for dir in $(go list ./... | grep -v /vendor/); do
#      echo "Building test $dir"
#      go test -c -o $TEST_BIN_DIR/${dir////_} $dir
#  done
#
#  # Start a second script to run all the test binaries
#  run_tests_script=$(cat <<'END_SCRIPT'
#  #!/bin/bash
#  set -e
#
#  cd "$(dirname "$0")"
#
#  EXIT_CODE=0
#  for binary in ./*; do
#    if [[ $binary == *.sh ]]; then
#    continue
#    fi
#    echo "Running tests in $binary"
#    if ! $binary; then
#      EXIT_CODE=1
#    fi
#  done
#  exit $EXIT_CODE
#END_SCRIPT
#  )
#
#  echo "$run_tests_script" > ${TEST_BIN_DIR}run_tests.sh
#  chmod +x ${TEST_BIN_DIR}run_tests.sh
#fi
#
## Run all tests if passed "--tests" argument
#if [ "$run_tests" == "true" ]; then
#  echo "Running all tests..."
#  if ! ${TEST_BIN_DIR}run_tests.sh; then
#    echo "Tests failed"
#    exit 1
#  else
#    echo "Tests passed"
#  fi
#fi
