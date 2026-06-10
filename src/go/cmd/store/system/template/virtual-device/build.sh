#!/usr/bin/env sh

# Build docker image
docker build -t {{ container-registry }}/m2cp-virtual-device-store-{{ store-name }}:latest .
