#!/usr/bin/env bash
/root/chain33 -f /root/chain33.toml &
sleep 10
/root/chain33 -f "$PARAFILE"
