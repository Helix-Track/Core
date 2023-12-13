#!/bin/bash

HERE=$(pwd)

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

SCRIPT_GET_JQ="$SUBMODULES_HOME/Software-Toolkit/Utils/Sys/Programs/get_jq.sh"

if ! test -e "$SCRIPT_GET_JQ"; then

    echo "ERROR: Script not found '$SCRIPT_GET_JQ'"
    exit 1
fi

MESSAGE="Hello!"
BOT_SCRIPT="$HERE/Run/Include/Curl/telegram_bot.sh"

if ! test -e "$BOT_SCRIPT"; then

    echo "ERROR: Script not found '$BOT_SCRIPT'"
    exit 1
fi

if [ -n "$1" ]; then

    MESSAGE="$1"
fi

if [ -z "$TELEGRAM_BOT" ]; then

    echo "ERROR: The 'TELEGRAM_BOT' env. variable is not defined"
    exit 1
fi

if [ -z "$TELEGRAM_BOT_TOKEN" ]; then

    echo "ERROR: The 'TELEGRAM_BOT_TOKEN' env. variable is not defined"
    exit 1
fi

if [ -z "$TELEGRAM_CHAT_ID" ]; then

    echo "ERROR: The 'TELEGRAM_CHAT_ID' env. variable is not defined"
    exit 1
fi

if bash "$SCRIPT_GET_JQ" >/dev/null 2>&1; then

    ENCODED_MESSAGE=$(jq -rn --arg x "$MESSAGE" '$x|@uri')

    bash "$BOT_SCRIPT" "$TELEGRAM_BOT" "$TELEGRAM_BOT_TOKEN" "$TELEGRAM_CHAT_ID" "$ENCODED_MESSAGE"

else

    echo "ERROR: JQ not available"
    exit 1
fi