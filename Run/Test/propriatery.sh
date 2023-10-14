#!/bin/bash

HERE="$(pwd)"

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

sh "$SUBMODULES_HOME/Testable/test.sh" "$HERE/Recipes" "$HERE/Propriatery"