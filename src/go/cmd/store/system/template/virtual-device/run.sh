#!/usr/bin/env sh

echo "WARNING: This script is only intended to test the image. For starting a virtual device properly, please use the m2cp CLI tool!"
docker rm --force VD && docker run -d --restart unless-stopped --add-host=host.docker.internal:host-gateway --privileged -p 30000:5555 -p 30001:5672 --name VD {{ container-registry }}/m2cp-virtual-device-store-{{ store-name }}:latest
