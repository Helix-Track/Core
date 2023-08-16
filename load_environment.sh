#!/bin/bash

HERE="$(pwd)"
SUBMODULES_PATH="$HERE/_Submodules"

if test -e "$SUBMODULES_PATH"; then

    echo "Exporting the submodules path: $SUBMODULES_PATH"
    
    export SUBMODULES_HOME="$SUBMODULES_PATH"
fi
