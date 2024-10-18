#!/usr/bin/env bash
LOG_FOLDER="log"

DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR" && cd ..

newest_log=$(cd "./$LOG_FOLDER" && ls -tp | grep -v /$ | head -1)
log_loc="./$LOG_FOLDER/$newest_log"
echo "Tailing \`$log_loc\` ..."
sleep 1s

tail "$log_loc" -f
