#!/bin/bash

if [ -n "$1" ]; then

  HERE="$(pwd)"

  cd "$HERE" && sh Core/Toolkit/Utils/Docker/get_container_terminal.sh "postgres.$1"

else

  echo "ERROR: Database name parameter not provided"
  exit 1
fi