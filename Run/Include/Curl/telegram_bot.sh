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

