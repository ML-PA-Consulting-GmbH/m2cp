#!/bin/bash

# Get exact path to the systemctl binary to call it directly
systemctl="$(command -v systemctl)"

# Refresh apt lists, if they do not exist yet
if [ ! -e /var/lib/apt/lists ]; then
    apt-get update
fi

# Enable required systemd services
"$systemctl" enable init.service
"$systemctl" enable ssh.service
"$systemctl" enable snapd.apparmor

# Set root password to 'm2cp'
chpasswd <<<"root:m2cp"

if grep -q securityfs /proc/filesystems; then
    mount -o rw,nosuid,nodev,noexec,relatime securityfs -t securityfs /sys/kernel/security
fi
mount -t tmpfs tmpfs /run

# Starts optional socat-calls depending on the store. E.g. in case of local store it remaps  
# snapstore url of localhost:3000 to host.docker.internal:3000
/bin/bash /bin/network-relay.sh

# Ensures that all snap service snaps (=snaps that are a daemon service) are running
bash /usr/local/bin/snap-daemon-starter.sh &

# Start systemd with init as first service
exec /lib/systemd/systemd --system init.service
