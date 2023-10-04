#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

HERE="$(pwd)"

cd "$HERE/Core" && sh "$SUBMODULES_HOME/Dependable/install_dependencies.sh"