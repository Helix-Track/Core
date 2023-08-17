#!/bin/bash

if [ -n "$1" ]; then

    echo "Executing post push steps for upstreams from $1"
fi

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

SCRIPT_GATHER_SUBMODULES="$SUBMODULES_HOME/Software-Toolkit/Utils/Git/gather_submodules.sh"

if test -e "$SCRIPT_GATHER_SUBMODULES"; then

  sh "$SCRIPT_GATHER_SUBMODULES"

else

  echo "ERROR: Script not found '$SCRIPT_GATHER_SUBMODULES'"
  exit 1
fi