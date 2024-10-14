#!/usr/bin/env bash

# Get the current directory of this script and the script name then cd to it
DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_NAME=$(basename "$0")
cd "$DIR"

declare -a profiles=("heap.svg" "allocs.svg")

clean() {
  # Clean up old performance profiles charting images
  for i in "${profiles[@]}"; do
    if [[ -f "$i" ]]; then
      echo "[$SCRIPT_NAME] $i exists; removing $i ..."
      rm --force "$i"
      if [[ ! -f "$i" ]]; then
        echo "[$SCRIPT_NAME] $i successfully DELETED"
      fi
    fi
  done
}

generate() {
  echo "[$SCRIPT_NAME] Generating performance profiles ..."
  for i in "${profiles[@]}"; do
    go tool pprof -svg http://localhost:6060/debug/pprof/heap > "$i"
  done
}

check() {
  # Check to make sure they exist after generating performance profiles charts
  for i in "${profiles[@]}"; do
    if [[ -f "$i" ]]; then
      echo "[$SCRIPT_NAME] Successfully generated: $i"
    else
      echo "[$SCRIPT_NAME] Failed to generate: $i"
    fi
  done
}

# do the work
clean
generate
check