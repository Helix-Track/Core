#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

DIR_TOOLKIT="$SUBMODULES_HOME/Software-Toolkit"

bash "$DIR_TOOLKIT/Utils/SonarQube/configure_sonar_lint.sh"