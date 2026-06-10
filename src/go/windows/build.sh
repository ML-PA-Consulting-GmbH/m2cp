#!/bin/bash

set -e  # exit on error
set -u  # exit on undefined variables
set -o pipefail  # exit on command pipe failures

cd "$(dirname "$0")"

INSTALLER_TEMPLATE="installer.template.iss"
INSTALLER_CUSTOMIZED="installer.iss"

clean() {
    rm -rf build/* && mkdir -p build/ && chmod 777 build
    rm -rf stage/*
    rm -f ${INSTALLER_CUSTOMIZED}
}

prepare() {
    mkdir -p ./stage/ && cp ../build/win-amd64/m2cp.exe stage/
    # TODO: better extract from the Debian package?
    echo "Downloading docs:"
    curl -X GET http://10.0.1.28:7000/m2cp/docs-html/latest/meta | jq
    mkdir -p ./stage/docs && curl -X GET http://10.0.1.28:7000/m2cp/docs-html/latest | tar -xz -C ./stage/docs
    # Remove Windows Zone.Identifier files created when downloading from web
    find ./stage/docs -name "*Zone.Identifier" -delete
}

build() {
    # customize installer: add version number
    VERSION=${BUILD_VERSION}
    cp ${INSTALLER_TEMPLATE} ${INSTALLER_CUSTOMIZED}
    sed -i 's/#define MyAppVersion "PLACEHOLDER"/#define MyAppVersion "'"${VERSION}"'"/' ${INSTALLER_CUSTOMIZED}

    docker run --rm -i -v $PWD:/work amake/innosetup ./${INSTALLER_CUSTOMIZED}
    rm ${INSTALLER_CUSTOMIZED}

    # the output file is owned by some other user do to usage of docker - we change the ownership by copying ind overwriting it
    cp build/m2cp-setup.exe /tmp/m2cp-setup.exe
    ls -la build
    chmod 755 build
    ls -la build
    mv /tmp/m2cp-setup.exe build/
}

clean
prepare
build
