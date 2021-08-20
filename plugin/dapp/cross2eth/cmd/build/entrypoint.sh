#!/usr/bin/env bash
/root/chain33 -f /root/chain33.toml &
# to wait nginx start
sleep 15
/root/chain33 -f "$PARAFILE"
