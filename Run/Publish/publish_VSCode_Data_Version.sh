#!/bin/bash

HERE=$(pwd)

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

BOT_SCRIPT="$HERE/Run/Include/Curl/telegram_bot.sh"
SCRIPT_GET_JQ="$SUBMODULES_HOME/Software-Toolkit/Utils/Sys/Programs/get_jq.sh"
SCRIPT_PUBLISH="$SUBMODULES_HOME/Software-Toolkit/Utils/VSCode/publish_new_data_version.sh"

if ! test -e "$SCRIPT_PUBLISH"; then

  echo "ERROR: Script not found '$SCRIPT_PUBLISH'"
  exit 1
fi

if sh "$SCRIPT_PUBLISH" "$1"; then

  if [ -n "$TELEGRAM_BOT" ] && [ -n "$TELEGRAM_BOT_TOKEN" ] && [ -n "$TELEGRAM_CHAT_ID" ]; then

    if ! test -e "$BOT_SCRIPT"; then

      echo "ERROR: Script not found '$BOT_SCRIPT'"
      exit 1
    fi

    if ! test -e "$SCRIPT_GET_JQ"; then

        echo "ERROR: Script not found '$SCRIPT_GET_JQ'"
        exit 1
    fi

    if sh "$SCRIPT_GET_JQ" >/dev/null 2>&1; then

        MESSAGE="The new version of the VSCode Data has been published"

        if [ -n "$VSCODE_DATA_PUBLISH_DESTINATION" ]; then

          FILE_DATA_VERSION_NAME="data_version.txt"
          FILE_DATA_VERSION="$VSCODE_DATA_PUBLISH_DESTINATION/$FILE_DATA_VERSION_NAME"

          if test -e "$FILE_DATA_VERSION"; then

            VERSION=$(cat "$FILE_DATA_VERSION")
            
            MESSAGE="$MESSAGE: $VERSION"
          fi
        fi

        ENCODED_MESSAGE=$(jq -rn --arg x "$MESSAGE" '$x|@uri')

        sh "$BOT_SCRIPT" "$TELEGRAM_BOT" "$TELEGRAM_BOT_TOKEN" "$TELEGRAM_CHAT_ID" "$ENCODED_MESSAGE"

    else

        echo "ERROR: JQ not available"
        exit 1
    fi
  fi
fi