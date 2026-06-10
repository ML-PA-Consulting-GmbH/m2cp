#!/bin/bash
# Recreate the initial environment from docker run
$(export -p)

# Force these environment variables
export PATH="/snap/bin:/usr/bin:/bin:/usr/sbin:/sbin"

# Mount the rabbitmq config into the rabbitmq snap
mount --bind -o nodev,ro /root/rabbitmq-env.conf /snap/m2cp-message-hub/current/etc/rabbitmq/rabbitmq-env.conf

systemctl enable --now ssh
snap restart m2cp-message-hub
