#!/bin/bash

if [ -z "$1" ]; then

    echo "ERROR: The Bot name parameter is mandatory"
fi

if [ -z "$2" ]; then

    echo "ERROR: The Bot token parameter is mandatory"
fi

if [ -z "$4" ]; then

    echo "ERROR: Chat ID parameter is mandatory"
fi

if [ -z "$4" ]; then

    echo "ERROR: Message parameter is mandatory"
fi

BOT="$1"
TOKEN="$2"
CHAT_ID="$3"
MESSAGE="$4"

DIR="$(dirname "$0")"

SCRIPT_API_CALL="$DIR/api_call.sh"

if test -e "$SCRIPT_API_CALL"; then

  # shellcheck disable=SC1090
  . "$SCRIPT_API_CALL"

else

  echo "ERROR: '$SCRIPT_API_CALL' not found"
  exit 1
fi

echo "We are about to trigger the API for the Telegram Bot: $BOT"

PORT="80"
HOST="api.telegram.org"

API_CALL "https" "$HOST" "$PORT" "bot$TOKEN/sendMessage?chat_id=$CHAT_ID&text=$MESSAGE"
