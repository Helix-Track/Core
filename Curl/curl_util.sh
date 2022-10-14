#!/bin/bash

# TODO: Move into the commonly used utility!

function RUN_CURL {

  CURL_UTIL_PORT="80"
  CURL_UTIL_ENDPOINT=""
  CURL_UTIL_HOST="127.0.0.1"
  CURL_UTIL_PROTOCOL="https"

  if [ -n "$1" ]; then

    CURL_UTIL_PROTOCOL="$1"
  fi

  if [ -n "$2" ]; then

    CURL_UTIL_HOST="$2"
  fi

  if [ -n "$3" ]; then

    CURL_UTIL_PORT="$3"
  fi

  if [ -n "$4" ]; then

    CURL_UTIL_ENDPOINT="$4"
  fi

  TARGET="$CURL_UTIL_PROTOCOL"://"$CURL_UTIL_HOST":"$CURL_UTIL_PORT"/"$CURL_UTIL_ENDPOINT"
  clear && echo "URL: $TARGET" && curl -v "$TARGET" && echo ""
}