#!/bin/bash

CURL="curl_util.sh"

if test -e "$CURL"; then

  . curl_util.sh

else

  echo "ERROR: '$CURL' not found"
  exit 1
fi

RUN_CURL "$1" "$2" "$3" "$4"