#!/usr/bin/env bash
PERF_FOLDER="perf"

# Get the current directory of this script
DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the script name then cd to it, but go one directory above /scripts
SCRIPT_NAME=$(basename "$0")
cd "$DIR" && cd ..

PERF_DIR="$PWD/$PERF_FOLDER"
declare -a profiles=("heap" "allocs")

create_dir() {
  if [[ -d "$PERF_DIR" ]]; then
    echo "[$SCRIPT_NAME] \`$PERF_DIR\` exists; skipping folder creation ..."
  else
    echo "[$SCRIPT_NAME] \`$PERF_DIR\` does not exist; creating \`$PERF_DIR\` ..."
    mkdir "$PERF_DIR"
  fi
}

clean() {
  # Clean up old performance profiles charting images
  for i in "${profiles[@]}"; do
    if [[ -f "$PERF_DIR/$i.svg" ]]; then
      echo "[$SCRIPT_NAME] \`$i.svg\` exists; removing \`$i.svg\` ..."
      rm --force "$PERF_DIR/$i.svg"

      if [[ ! -f "$PERF_DIR/$i.svg" ]]; then
        echo "[$SCRIPT_NAME] \`$i.svg\` successfully DELETED"
      fi
    fi
  done
}

generate() {
  echo "[$SCRIPT_NAME] Generating performance profiles with \`go tool pprof\`..."
  for i in "${profiles[@]}"; do
    go tool pprof -svg http://localhost:6060/debug/pprof/"$i" > "$PERF_DIR/$i.svg"
  done
}

check() {
  # Check to make sure they exist after generating performance profiles charts
  for i in "${profiles[@]}"; do
    if [[ -f "$PERF_DIR/$i.svg" ]]; then
      echo "[$SCRIPT_NAME] Successfully generated: \`$i.svg\`"
    else
      echo "[$SCRIPT_NAME] Failed to generate: \`$i.svg\`"
    fi
  done
}

# do the work
create_dir
clean
generate
check
echo ""