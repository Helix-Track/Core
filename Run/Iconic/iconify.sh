#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

SCRIPT_ICONIFY="$SUBMODULES_HOME/Iconic/iconify.sh"

if ! test -e "$SCRIPT_ICONIFY"; then

    echo "ERROR: Script not found '$SCRIPT_ICONIFY'"
    exit 1
fi

PARAMS=""

if [ -n "$1" ]; then

  PARAMS="$1"
fi

sh "$SCRIPT_ICONIFY" "$PARAMS"

