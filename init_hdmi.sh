#!/usr/bin/env bash

set -e

# Logging function
log() {
    echo "[kindavm-init] $*" >&2
}

log "Starting KindaVM HDMI initialization"

DEVICE=/dev/video0
V4L2_CTL=/usr/bin/v4l2-ctl
EDID_FILE="/usr/local/lib/kindavm/edid.hex"

log "Setting EDID from $EDID_FILE"
$V4L2_CTL -d $DEVICE --set-edid=file=$EDID_FILE

sleep 3
log "Querying DV timings"
$V4L2_CTL -d $DEVICE --query-dv-timings
log "Setting DV timings"
$V4L2_CTL -d $DEVICE --set-dv-bt-timings query

log "HDMI initialization complete"
