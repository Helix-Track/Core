#!/bin/bash

DIR_HOME="$(readlink --canonicalize ~)"

if [ -z "$SHARES_SERVER" ]; then

    echo "ERROR: 'SHARES_SERVER' variable is not set"
    exit 1
fi

# shellcheck disable=SC2034
DIR_INSTALLATION_HOME="$DIR_HOME/Yandex.Disk/Workspace/VSCode/Linux/VSCode"

# shellcheck disable=SC2034
DOWNLOAD_URL_EXTENSIONS="http://$SHARES_SERVER:8081/extensions.1.0.0.tar.gz"

# shellcheck disable=SC2034
DOWNLOAD_URL_USER_DATA="http://$SHARES_SERVER:8081/user-data.1.0.0.tar.gz"

