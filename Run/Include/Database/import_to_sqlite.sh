#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined !!!"
  exit 1
fi

sh "$SUBMODULES_HOME/Software-Toolkit/Utils/Db/import_to_sqlite.sh Database/Definition.sqlite" "$1"