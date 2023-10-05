#!/bin/bash

HERE="$(pwd)"

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

cd "$HERE/Core" && sh "$SUBMODULES_HOME/Installable/install.sh"