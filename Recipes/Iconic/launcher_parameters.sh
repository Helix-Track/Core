#!/bin/bash

HERE=$(pwd)

# shellcheck disable=SC2034
BIN="$HERE/open"

# shellcheck disable=SC2034
LAUNCHER="$HERE/Assets/Launcher.svg"

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
NAME="$VERSIONABLE_NAME"

# shellcheck disable=SC2034
DESCRIPTION="$VERSIONABLE_DESCRIPTION"




# TODO: Re-organise Recipes by adding proper subdirectories