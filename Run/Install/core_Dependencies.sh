#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

sh "$SUBMODULES_HOME/Dependable/install_dependencies.sh"