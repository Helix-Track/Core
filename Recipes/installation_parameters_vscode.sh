#!/bin/bash

DIR_HOME="$(readlink --canonicalize ~)"

if [ -z "$SHARES_SERVER" ]; then

    echo "ERROR: 'SHARES_SERVER' variable is not set"
    exit 1
fi

# shellcheck disable=SC2034
DATA_VERSION="1.0.9"

# shellcheck disable=SC2034
DIR_INSTALLATION_HOME="$DIR_HOME/Workspaces/VSCode/Linux/VSCode"

# shellcheck disable=SC2034
DOWNLOAD_URL_EXTENSIONS="http://$SHARES_SERVER:8081/extensions.$DATA_VERSION.tar.gz"

# shellcheck disable=SC2034
DOWNLOAD_URL_USER_DATA="http://$SHARES_SERVER:8081/user-data.$DATA_VERSION.tar.gz"

