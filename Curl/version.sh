#!/bin/bash

CURL="../Core/Toolkit/Utils/curl.sh"

if test -e "$CURL"; then

  # shellcheck disable=SC1090
  . "$CURL"

else

  echo "ERROR: '$CURL' not found"
  exit 1
fi

RUN_CURL "$1" "$2" "$3" "$4"