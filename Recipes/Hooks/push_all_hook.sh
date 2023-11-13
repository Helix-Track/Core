#!/bin/bash

if [ -z "$1" ]; then

    echo "ERROR: Upstreams directory path is mandatory"
fi

DIR_UPSTREAMS="$1"
DIR_ROOT="$DIR_UPSTREAMS/.."
DIR_GIT="$DIR_ROOT/.git"

if ! test -e "$DIR_GIT"; then

    echo "ERROR: Not a Git root directory '$DIR_GIT'"
    exit 1
fi

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

SCRIPT_GATHER_SUBMODULES="$SUBMODULES_HOME/Software-Toolkit/Utils/Git/gather_submodules.sh"
SCRIPT_PUSH_SUBMODULES="$SUBMODULES_HOME/Software-Toolkit/Utils/Git/push_all_submodules_head_positions.sh"

if ! test -e "$SCRIPT_GATHER_SUBMODULES"; then

    echo "ERROR: Script not found '$SCRIPT_GATHER_SUBMODULES'"
    exit 1    

else

    if ! test -e "$SCRIPT_PUSH_SUBMODULES"; then

        echo "ERROR: Script not found '$SCRIPT_PUSH_SUBMODULES'"
        exit 1    
    fi

    sh "$SCRIPT_GATHER_SUBMODULES" && sh "$SCRIPT_PUSH_SUBMODULES" "$DIR_ROOT"
fi