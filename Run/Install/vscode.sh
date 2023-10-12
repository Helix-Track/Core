#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

HERE="$(pwd)"
DIR_TOOLKIT="$SUBMODULES_HOME/Software-Toolkit"

sh "$DIR_TOOLKIT/Utils/VSCode/install.sh" "$HERE/Recipes/installation_parameters_vscode.sh"