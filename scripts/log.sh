#!/usr/bin/env bash
DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

newest_log=$(cd "./log" && ls -tp | grep -v /$ | head -1)
log_loc="./log/$newest_log"
echo "Tailing $log_loc ..."
sleep 2s

tail "$log_loc" -f