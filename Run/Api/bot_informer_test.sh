#!/bin/bash

HERE=$(pwd)

MESSAGE="Hello from '$USER'"
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

sh "$BOT_SCRIPT" "$TELEGRAM_BOT" "$TELEGRAM_BOT_TOKEN" "$TELEGRAM_CHAT_ID" "$MESSAGE"