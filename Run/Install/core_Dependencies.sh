#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

bash "$SUBMODULES_HOME/Dependable/install_dependencies.sh"