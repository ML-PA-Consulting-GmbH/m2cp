#!/bin/bash

# ROOT permission is required to run this script
# This script adds or removes an IP address to/from the loopback interface

# Usage: ./ip-address-test.sh <add|remove> <ip_address>
# Example: ./ip-address-test.sh add "fd00::100/128"

# Check if the correct number of arguments is provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <add|remove> <ip_address>"
    exit 1
fi

# Get the action and IP address from the arguments
ACTION=$1
IP_ADDRESS=$2
# Check if the action is valid
if [ "$ACTION" != "add" ] && [ "$ACTION" != "remove" ]; then
    echo "Invalid action: $ACTION. Use 'add' or 'remove'."
    exit 1
fi


# Perform the action
if [ "$ACTION" == "add" ]; then
    # Add the IP address to the loopback interface
    sudo ip addr add "$IP_ADDRESS" dev lo
    echo "Added IP address $IP_ADDRESS to loopback interface."
elif [ "$ACTION" == "remove" ]; then
    # Remove the IP address from the loopback interface
    sudo ip addr del "$IP_ADDRESS" dev lo
    echo "Removed IP address $IP_ADDRESS from loopback interface."
fi