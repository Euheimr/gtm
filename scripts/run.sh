#!/usr/bin/env bash
export CGO_ENABLED=0

BINARY_NAME="gtm"
MAIN_GO_FOLDER="cmd"
BINARY_FOLDER="bin"

# this is a local variable that's set to 1 in case a build was attempted but failed
build_attempted=0

# Get the current directory of this script and the script name then cd to it
DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_NAME=$(basename "$0")
cd "$DIR"

echo "[$SCRIPT_NAME] CGO_ENABLED=$CGO_ENABLED"

# This is meant to rename the binary with the customary windows .EXE if we are building on windows
BINARY_POSTFIX=""
if [[ $OSTYPE == "msys" || $OSTYPE == "cygwin" ]]; then
  echo "[$SCRIPT_NAME] Operating system is windows... setting binary postfix to \`.exe\` ..."
  BINARY_POSTFIX=".exe"
fi

BINARY_NAME=$BINARY_NAME$BINARY_POSTFIX
BINARY_PATH="$BINARY_FOLDER/$BINARY_NAME"

clean () {
  if [[ -d "$BINARY_FOLDER" ]]; then
    echo "[$SCRIPT_NAME] \`$PWD/$BINARY_FOLDER\` already exists. Skipping folder creation ..."
    if [[ -f "$BINARY_PATH" ]]; then
      echo "[$SCRIPT_NAME] Removing old binary: \`$PWD/$BINARY_PATH\` ..."
      rm --force "$BINARY_PATH"
      if [[ ! -f "$BINARY_PATH" ]]; then
        echo "[$SCRIPT_NAME] Successfully DELETED $PWD/$BINARY_PATH !"
      fi
    else
      echo "[$SCRIPT_NAME] No binaries to cleanup at \`$PWD/$BINARY_PATH\` ... "
    fi
  else
    echo "[$SCRIPT_NAME] \`$PWD/$BINARY_FOLDER\` does not exist! Creating \`$PWD/$BINARY_FOLDER\` ..."
    mkdir "$BINARY_FOLDER"

    if [[ ! -d "$BINARY_FOLDER" ]]; then
      echo "[$SCRIPT_NAME] ERROR: Failed to create $BINARY_FOLDER!"
    fi
  fi
}

build (){
  clean
  echo "[$SCRIPT_NAME] Building binary at \`$PWD/$BINARY_PATH\` ..."
  go build -o "$BINARY_PATH" "$MAIN_GO_FOLDER/main.go"

  build_attempted=1

  if [[ -f "$BINARY_PATH" ]]; then
    echo "[$SCRIPT_NAME] BUILD SUCCESS !";
    sleep 2s
  else
    echo "[$SCRIPT_NAME] ERROR: BUILD FAILED to compile !!!"
    sleep 2s
    exit
  fi
}

run (){
  if [[ -f "$BINARY_PATH" ]]; then
    echo "[$SCRIPT_NAME] Running binary \`$BINARY_NAME\` at \`$BINARY_FOLDER\` ..."
    "$BINARY_PATH"
  else
    echo "[$SCRIPT_NAME] Could not find binary at \`$BINARY_PATH\` !"
  fi
}

####     It's business time!     ####
while [ $# -gt 0 ]; do
  # Process option flags passed to ./run.sh ...
  case $1 in
    -b | build | --build ) build
      ;;
    -c | clean | --clean ) clean; exit;
      ;;
    # this is a "catch-all" option flag
    *) echo "[$SCRIPT_NAME] ERROR Invalid Option: \`${1}\`!";
    echo "[$SCRIPT_NAME] Use \`-b\`, \`build\`, or \`--build\` to force a build. OR \`-c\`, \`clean\`, or \`--clean\` to clear out any binaries in /bin !";
    exit
      ;;
  esac
  shift
done

# At this point, if the binary exists and no flags were passed - just run it
if [[ -f $BINARY_PATH ]]; then
  run
else
  # If a build was not attempted and the binary doesn't exist, attempt a build & run
  if [[ ${build_attempted} = 0 ]]; then
  echo "[$SCRIPT_NAME] Could not find binary. Attempting a build ..."
  build
  run
  fi
fi
