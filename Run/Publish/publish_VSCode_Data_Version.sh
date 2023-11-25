#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

SCRIPT_PUBLISH="$SUBMODULES_HOME/Software-Toolkit/Utils/VSCode/publish_new_data_version.sh"

if ! test -e "$SCRIPT_PUBLISH"; then

    echo "ERROR: Script not found '$SCRIPT_PUBLISH'"
    exit 1
fi

sh "$SCRIPT_PUBLISH" "$1"