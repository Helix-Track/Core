#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined !!!"
  exit 1
fi

bash "$SUBMODULES_HOME/Software-Toolkit/Utils/Db/import_to_postgres.sh" Definition "$1" "postgres" "Test12345" _Databases/Postgres