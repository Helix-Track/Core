#!/bin/bash

HERE="$(pwd)"

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

cd "$HERE/Propriatery" && sh "$SUBMODULES_HOME/Dependable/install_dependencies.sh"