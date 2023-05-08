#!/bin/bash

if [ -n "$1" ]; then

  HERE="$(pwd)"

  cd "$HERE" && sh Core/Toolkit/Utils/Docker/get_container_terminal.sh "sonarqube.$1"

else

  echo "ERROR: SonarQube name parameter not provided"
  exit 1
fi