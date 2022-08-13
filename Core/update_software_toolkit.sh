#!/bin/bash

HERE="$(pwd)"
DIR_DEPENDABLE="Dependable"
DIR_VERSIONABLE="Versionable"
DIR_INSTALLABLE="Installable"

function UPDATE_MODULE {

  if [ -z "$1" ]; then

    echo "ERROR: The Software Toolkit module directory not provided"
    exit 1
  fi

  DIR="$1"

  if test -e "$DIR"; then

    if cd "$DIR" && git checkout main && git fetch && git pull && cd "$HERE"; then

      cd "$HERE" &&
        echo "The Software Toolkit module updated: '$DIR'"
    else

      echo "ERROR: Could not update the Software Toolkit: '$DIR'"
      exit 1
    fi
  else

    echo "WARNING: Software Toolkit module '$DIR' not found at '$(pwd)'"
  fi
}

UPDATE_MODULE "$DIR_DEPENDABLE"
UPDATE_MODULE "$DIR_VERSIONABLE"
UPDATE_MODULE "$DIR_INSTALLABLE"
