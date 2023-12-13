#!/bin/bash

HERE=$(pwd)

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

DIR_TOOLKIT="$SUBMODULES_HOME/Software-Toolkit"
RECIPE_SETTINGS_JSON="$HERE/Recipes/VSCode/settings.json.sh"
SCRIPT_GET_PROGRAM="$DIR_TOOLKIT/Utils/Sys/Programs/get_program.sh"
SCRIPT_EXTEND_JSON="$DIR_TOOLKIT/Utils/Sys/JSON/merge_jsons.sh"

if ! test -e "$SCRIPT_GET_PROGRAM"; then

  echo "ERROR: Script not found '$SCRIPT_GET_PROGRAM'"
  exit 1
fi

if ! test -e "$SCRIPT_EXTEND_JSON"; then

  echo "ERROR: Script not found '$SCRIPT_EXTEND_JSON'"
  exit 1
fi

if bash "$SCRIPT_GET_PROGRAM" code >/dev/null 2>&1; then

  CODE_PATH=$(which code)

  if test -e "$CODE_PATH"; then

    echo "VSCode path: '$CODE_PATH'"

  else

    echo "ERROR: VSCode Path not found '$CODE_PATH'"
    exit 1
  fi

  CODE_DIR=$(dirname "$CODE_PATH")
  SETTINGS_JSON="$CODE_DIR/data/user-data/User/settings.json"

  echo "Checking: '$SETTINGS_JSON'"

  if test -e "$SETTINGS_JSON"; then

    echo "Settings JSON: '$SETTINGS_JSON'"

  else

    if echo "{}" >> "$SETTINGS_JSON"; then

      echo "Settings JSON created: '$SETTINGS_JSON'"

    else

      echo "ERROR: Could not create '$SETTINGS_JSON'"
      exit 1
    fi
  fi

  if test -e "$RECIPE_SETTINGS_JSON"; then

    echo "Settings JSON recipe: '$RECIPE_SETTINGS_JSON'"

  else

    echo "ERROR: Settings JSON recipe not found '$RECIPE_SETTINGS_JSON'"
    exit 1
  fi

  # shellcheck disable=SC1090
  if bash "$SCRIPT_EXTEND_JSON" "$SETTINGS_JSON" "$RECIPE_SETTINGS_JSON" "$SETTINGS_JSON"; then

    echo "Settings JSON has been extended with success"

  else

    echo "ERROR: Settings JSON failed to extend"
    exit 1
  fi
fi