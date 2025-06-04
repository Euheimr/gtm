#!/bin/bash
export CGO_ENABLED=0

BINARY_NAME="gtm"
MAIN_GO_FOLDER="cmd"
BINARY_FOLDER="bin"
SCRIPT_NAME="${0%/}"              # strip trailing slash
SCRIPT_NAME="${SCRIPT_NAME##*/}"  # get the script name

# this is a local variable that's set to 1 in case a build was attempted but failed
build_attempted=0
binary_postfix=""

# Get the current directory of this script
DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIR="${DIR%/}"  # strip trailing slash
cd "${DIR}"

# getting the last folder is MUCH faster this way instead of using $(basename ...) cmd
SUBDIR="${DIR##*/}"

if [ "${SUBDIR}" = "scripts" ]; then
    # go one directory above scripts/
    cd ..
fi

# This is meant to rename the binary with the customary windows .EXE only if we are
#   building on windows.
if [ "${OS}" = "Windows_NT" ]; then
  binary_postfix=".exe"
fi

BINARY_NAME=${BINARY_NAME}${binary_postfix}
BINARY_PATH="${BINARY_FOLDER}/${BINARY_NAME}"

clean () {
  if [ -d "${BINARY_FOLDER}" ]; then
    echo "[${SCRIPT_NAME}] \`${PWD}/${BINARY_FOLDER}\` already exists. Skipping folder creation ..."
    if [ -f "${BINARY_PATH}" ]; then
      echo "[${SCRIPT_NAME}] Removing old binary: \`${PWD}/${BINARY_PATH}\` ..."
      rm -f "${BINARY_PATH}"
      if [ ! -f "${BINARY_PATH}" ]; then
        echo "[${SCRIPT_NAME}] DELETED ${PWD}/${BINARY_PATH}"
      fi
    else
      echo "[${SCRIPT_NAME}] No binaries to cleanup at \`${PWD}/${BINARY_PATH}\` ... "
    fi
  else
    echo "[${SCRIPT_NAME}] \`${PWD}/${BINARY_FOLDER}\` does not exist! Creating \`${PWD}/${BINARY_FOLDER}\` ..."
    mkdir "${BINARY_FOLDER}"

    if [ ! -d "${BINARY_FOLDER}" ]; then
      echo "[${SCRIPT_NAME}] ERROR: Failed to create ${BINARY_FOLDER}!"
    fi
  fi
}

build (){
  clean
  echo "[${SCRIPT_NAME}] Building binary at \`${PWD}/${BINARY_PATH}\` ..."
  go build -o "${BINARY_PATH}" "${MAIN_GO_FOLDER}/main.go"
  build_attempted=1

  if [ -f "${BINARY_PATH}" ]; then
    echo "[${SCRIPT_NAME}] BUILD SUCCESS !";
  else
    echo "[${SCRIPT_NAME}] ERROR: BUILD FAILED to compile !!!"
    sleep 10s
  fi
}

run (){
  if [ -f "${BINARY_PATH}" ]; then
    echo "[${SCRIPT_NAME}] Running binary \`${BINARY_NAME}\` at \`${BINARY_FOLDER}\` ..."
    sleep 2s
    "${BINARY_PATH}"
  else
    echo "[${SCRIPT_NAME}] Could not find binary at \`${BINARY_PATH}\` !"
  fi
}

################################################################################
#########################     It's business time!     ##########################
echo "[${SCRIPT_NAME}] CGO_ENABLED=${CGO_ENABLED}"
while [ $# -gt 0 ]; do
  # Process option flags passed to ./run.sh ...
  case $1 in
    -b | build | --build ) build
      ;;
    -c | clean | --clean ) clean; exit;
      ;;
    -bo | build-only | --build-only ) build;
       echo "[${SCRIPT_NAME}] Build done, exiting ...";
       exit;
      ;;
    # this is a "catch-all" option flag
    *) echo "[${SCRIPT_NAME}] ERROR Invalid Option flag: ${1}";
       echo "[${SCRIPT_NAME}] Use the following flags:";
       echo "    Force a Build & Run the executable with: --build, build, or -b";
       echo "    Force ONLY a build (DO NOT run ${BINARY_NAME}): --build-only, build-only, or -bo";
       echo "    Delete any binaries in /bin with: --clean, clean, or -c";
    exit
      ;;
  esac
  shift
done

# At this point, if the binary exists and no flags were passed; just run it
if [ -f ${BINARY_PATH} ]; then
  run
else
  # If a build was not attempted and the binary doesn't exist, attempt a build & run
  if [ ${build_attempted} = 0 ]; then
  echo "[${SCRIPT_NAME}] Could not find binary. Attempting a build ..."
  build
  run
  fi
fi

echo ""
