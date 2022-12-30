#!/bin/bash
DIR="$(dirname "$0")"

API_CALL="$DIR/api_call.sh"

if test -e "$API_CALL"; then

  # shellcheck disable=SC1090
  . "$API_CALL"

else

  echo "ERROR: '$API_CALL' not found"
  exit 1
fi

API_CALL "$UTIL_PROTOCOL" "$UTIL_HOST" "$UTIL_PORT" jwt_check