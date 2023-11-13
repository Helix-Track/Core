#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

HERE="$(dirname "$SUBMODULES_HOME")"

# shellcheck disable=SC2034
CMD_PATH="$HERE"

# shellcheck disable=SC2034
CMD="open"

# shellcheck disable=SC2034
LAUNCHER="$HERE/Assets/Launcher.svg"

# shellcheck disable=SC2034
DESKTOP_FILE_NAME="helixtrack_core_dev.desktop"

SCRIPT_VERSION="$HERE/Core/Version/version.sh"

if ! test -e "$SCRIPT_VERSION"; then

    echo "ERROR: Version file not found '$SCRIPT_VERSION'"
    exit 1
fi

# shellcheck disable=SC1090
. "$SCRIPT_VERSION"

if [ -z "$VERSIONABLE_VERSION_PRIMARY" ]; then

    echo "ERROR: 'VERSIONABLE_VERSION_PRIMARY' variable not set"
    exit 1
fi

if [ -z "$VERSIONABLE_VERSION_SECONDARY" ]; then

    echo "ERROR: 'VERSIONABLE_VERSION_SECONDARY' variable not set"
    exit 1
fi

if [ -z "$VERSIONABLE_VERSION_PATCH" ]; then

    echo "ERROR: 'VERSIONABLE_VERSION_PATCH' variable not set"
    exit 1
fi

if [ -z "$VERSIONABLE_NAME" ]; then

    echo "ERROR: 'VERSIONABLE_NAME' variable not set"
    exit 1
fi

if [ -z "$VERSIONABLE_DESCRIPTION" ]; then

    echo "ERROR: 'VERSIONABLE_DESCRIPTION' variable not set"
    exit 1
fi

# shellcheck disable=SC2034
VERSION="$VERSIONABLE_VERSION_PRIMARY.$VERSIONABLE_VERSION_SECONDARY.$VERSIONABLE_VERSION_PATCH"

# shellcheck disable=SC2034
NAME="$VERSIONABLE_NAME DEV"

# shellcheck disable=SC2034
DESCRIPTION="The $VERSIONABLE_NAME development IDE."

DESKTOP_ICON=true

if [ -n "$HELIXTRACK_LAUNCHER_DESKTOP" ]; then

    # shellcheck disable=SC2034
    DESKTOP_ICON="$HELIXTRACK_LAUNCHER_DESKTOP"
fi

# TODO: Remove all unused submodules
# TODO: Re-organise Recipes by adding proper subdirectories