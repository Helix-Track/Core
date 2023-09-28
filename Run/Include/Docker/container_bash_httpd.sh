#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

if [ -n "$1" ]; then

  HERE="$(pwd)"

  cd "$HERE" && sh "$SUBMODULES_HOME"/Software-Toolkit/Utils/Docker/get_container_terminal.sh "httpd.$1"

else

  echo "ERROR: HTTPD name parameter not provided"
  exit 1
fi