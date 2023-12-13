#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined !!!"
  exit 1
fi

bash "$SUBMODULES_HOME/Software-Toolkit/Utils/Sys/Programs/get_docker.sh" true
