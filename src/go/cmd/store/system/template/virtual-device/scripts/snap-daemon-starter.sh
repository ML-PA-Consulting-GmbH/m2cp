#!/bin/bash

MAX_TRIES=60  # check every 5 seconds
TRIES=0
while true; do
  if systemctl is-active snapd.service > /dev/null; then
    echo "snapd systemd service is running"
    break
  fi

  TRIES=$((TRIES+1))
  if (( TRIES > MAX_TRIES )); then
    echo "Failed to start snapd systemd service within timeout"
    exit 1
  fi

  sleep 5
done

# Start snap services, if enabled but not active
# retrieve the list of services and extract the lines for inactive services
INACTIVE_SERVICES=$(snap services | grep -E '^[^ ]+.+enabled  inactive')

# loop through the lines for inactive services
while read SERVICE_DETAILS; do
  # extract the name of the service
  SERVICE_NAME=$(echo $SERVICE_DETAILS | awk '{ print $1 }')

  # start the service
  snap start $SERVICE_NAME
done <<< "$INACTIVE_SERVICES"
