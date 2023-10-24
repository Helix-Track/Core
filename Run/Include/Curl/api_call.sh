#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

CURL="$SUBMODULES_HOME/Software-Toolkit/Utils/curl.sh"

if test -e "$CURL"; then

  # shellcheck disable=SC1090
  . "$CURL"

else

  echo "ERROR: '$CURL' not found"
  exit 1
fi

API_CALL() {

  if [ -n "$1" ]; then

    UTIL_PROTOCOL="$1"
  fi

  if [ -n "$2" ]; then

    UTIL_HOST="$2"
  fi

  if [ -n "$3" ]; then

    UTIL_PORT="$3"
  fi

  if [ -z "$4" ]; then

    echo "ERROR: The API call endpoint parameter is mandatory"
    exit 1
  fi

  RUN_CURL "$UTIL_PROTOCOL" "$UTIL_HOST" "$UTIL_PORT" "$4"
}

