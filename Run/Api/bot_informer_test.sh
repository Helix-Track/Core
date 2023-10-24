#!/bin/bash

HERE="$(pwd)"

MESSAGE="Hello from $USER"

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

cd "$HERE" && sh Run/Include/Curl/telegram_bot.sh "$TELEGRAM_BOT" "$TELEGRAM_BOT_TOKEN" "$MESSAGE"