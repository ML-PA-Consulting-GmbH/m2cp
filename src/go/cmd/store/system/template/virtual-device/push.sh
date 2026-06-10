#!/usr/bin/env sh

# Push docker image
docker push {{ container-registry }}/m2cp-virtual-device-store-{{ store-name }}:latest
