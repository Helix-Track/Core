#!/bin/bash

HERE=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
DIR_HOME="$(readlink --canonicalize ~)"

if [ -z "$SHARES_SERVER" ]; then

    echo "ERROR: 'SHARES_SERVER' variable is not set"
    exit 1
fi

SCRIPT_DEFAULTS="$HERE/defaults.sh"

if ! test -e "$SCRIPT_DEFAULTS"; then

    echo "ERROR: Script not found '$SCRIPT_DEFAULTS'"
    exit 1
fi

# shellcheck disable=SC1090
. "$SCRIPT_DEFAULTS"

if [ -z "$DEFAULT_DATA_VERSION" ]; then

    echo "The 'DEFAULT_DATA_VERSION' is not defined"
    exit 1
fi

# shellcheck disable=SC2034
DATA_VERSION="$DEFAULT_DATA_VERSION"

OBTAINED_DATA_VERSION=$(curl "http://$SHARES_SERVER:8081/data_version.txt")

# shellcheck disable=SC2002
if ! echo "$OBTAINED_DATA_VERSION" | grep "404 Not Found" >/dev/null 2>&1; then

    if ! [ "$DATA_VERSION" == "$OBTAINED_DATA_VERSION" ]; then

        echo "New data version is available: $OBTAINED_DATA_VERSION"
        echo "Current recipe data version is: $DATA_VERSION"

        DATA_VERSION="$OBTAINED_DATA_VERSION"
    fi
fi

# shellcheck disable=SC2034
DIR_INSTALLATION_HOME="$DIR_HOME/Workspaces/VSCode/Linux/VSCode"

# shellcheck disable=SC2034
DOWNLOAD_URL_EXTENSIONS="http://$SHARES_SERVER:8081/extensions.$DATA_VERSION.tar.gz"

# shellcheck disable=SC2034
DOWNLOAD_URL_USER_DATA="http://$SHARES_SERVER:8081/user-data.$DATA_VERSION.tar.gz"

