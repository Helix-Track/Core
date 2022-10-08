#!/bin/bash

HERE="$(pwd)"
VERSIONABLE_BUILD_SCRIPT="$HERE/Versionable/versionable_build.sh"
VERSIONABLE_INSTALL_SCRIPT="$HERE/Versionable/versionable_install.sh"

if ! test -e "$VERSIONABLE_BUILD_SCRIPT"; then

  echo "ERROR: The versionable build script not found at expected location: '$VERSIONABLE_BUILD_SCRIPT'"
  exit 1
fi

if ! test -e "$VERSIONABLE_INSTALL_SCRIPT"; then

  echo "ERROR: The versionable install script not found at expected location: '$VERSIONABLE_INSTALL_SCRIPT'"
  exit 1
fi

cd "$HERE" && sh "$VERSIONABLE_BUILD_SCRIPT" Application .. &&  sh "$VERSIONABLE_INSTALL_SCRIPT" Application
