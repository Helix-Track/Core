#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

VERSIONABLE_BUILD_SCRIPT="$SUBMODULES_HOME/Versionable/versionable_build.sh"
VERSIONABLE_INSTALL_SCRIPT="$SUBMODULES_HOME/Versionable/versionable_install.sh"

if ! test -e "$VERSIONABLE_BUILD_SCRIPT"; then

  echo "ERROR: The versionable build script not found at expected location: '$VERSIONABLE_BUILD_SCRIPT'"
  exit 1
fi

if ! test -e "$VERSIONABLE_INSTALL_SCRIPT"; then

  echo "ERROR: The versionable install script not found at expected location: '$VERSIONABLE_INSTALL_SCRIPT'"
  exit 1
fi

# TODO: Migration to Go Lang
#
# cd "$HERE" && sh "$VERSIONABLE_BUILD_SCRIPT" Application .. &&  sh "$VERSIONABLE_INSTALL_SCRIPT" Application
