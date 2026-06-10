#!/bin/bash

# Expose snapd socket to port 5555
socat TCP-LISTEN:5555,reuseaddr,fork UNIX-CLIENT:/run/snapd.socket &
