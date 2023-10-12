#!/bin/bash

if [ -z "$SONARQUBE_PROJECT" ]; then

    echo "ERROR: Environment variable not set 'SONARQUBE_PROJECT'"
    exit 1
fi

if [ -z "$SONARQUBE_PORT" ]; then

    echo "ERROR: Environment variable not set 'SONARQUBE_PORT'"
    exit 1
fi

if [ -z "$SONARQUBE_ADMIN_USERNAME" ]; then

    echo "ERROR: Environment variable not set 'SONARQUBE_ADMIN_USERNAME'"
    exit 1
fi

if [ -z "$SONARQUBE_ADMIN_PASSWORD" ]; then

    echo "ERROR: Environment variable not set 'SONARQUBE_ADMIN_PASSWORD'"
    exit 1
fi

ADMIN_USER="$SONARQUBE_ADMIN_USERNAME"
ADMIN_PASSWORD="$SONARQUBE_ADMIN_PASSWORD"

# shellcheck disable=SC2034
SONARQUBE_NAME="$SONARQUBE_PROJECT"

# shellcheck disable=SC2034
DB_USER="db$ADMIN_USER"

# shellcheck disable=SC2034
DB_PASSWORD="db$ADMIN_PASSWORD"

