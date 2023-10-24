#!/bin/bash

if [ -z "$1" ]; then

    echo "ERROR: The Bot name parameter is mandatory"
fi

if [ -z "$2" ]; then

    echo "ERROR: The Bot token parameter is mandatory"
fi

if [ -z "$3" ]; then

    echo "ERROR: Message parameter is mandatory"
fi

BOT="$1"
TOKEN="$2"
MESSAGE="$3"

DIR="$(dirname "$0")"

API_CALL="$DIR/api_call.sh"

if test -e "$API_CALL"; then

  # shellcheck disable=SC1090
  . "$API_CALL"

else

  echo "ERROR: '$API_CALL' not found"
  exit 1
fi

echo "We are about to trigger the API for the Telegram Bot: $BOT"

# TODO:
#
# API_CALL "$UTIL_PROTOCOL" "$UTIL_HOST" "$UTIL_PORT" version
