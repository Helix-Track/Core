#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

if [ -n "$1" ]; then

  HERE="$(pwd)"

  cd "$HERE" && bash "$SUBMODULES_HOME"/Software-Toolkit/Utils/Docker/get_container_terminal.sh "postgres.$1"

else

  echo "ERROR: Database name parameter not provided"
  exit 1
fi