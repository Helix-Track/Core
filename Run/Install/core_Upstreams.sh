#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

bash "$SUBMODULES_HOME/Upstreamable/install_upstreams.sh"