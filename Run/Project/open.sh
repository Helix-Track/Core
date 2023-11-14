#!/bin/bash

HERE=$(pwd)
PROJECT="$HERE"
DIR_HOME=$(eval echo ~"$USER")
FILE_ZSH_RC="$DIR_HOME/.zshrc"
FILE_BASH_RC="$DIR_HOME/.bashrc"

FILE_RC=""
    
if test -e "$FILE_ZSH_RC"; then

  FILE_RC="$FILE_ZSH_RC"

else

    if test -e "$FILE_BASH_RC"; then

      FILE_RC="$FILE_BASH_RC"

    else

      echo "ERROR: No '$FILE_ZSH_RC' or '$FILE_BASH_RC' found on the system"
      exit 1
    fi
fi

# shellcheck disable=SC1090
. "$FILE_RC" >/dev/null 2>&1

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: The SUBMODULES_HOME is not defined"
  exit 1
fi

SCRIPT_OPEN="$SUBMODULES_HOME/Project/open.sh"

if ! test -e "$SCRIPT_OPEN"; then

    echo "ERROR: Script not found '$SCRIPT_OPEN'"
    exit 1
fi

IDE="code"

if [ -n "$HELIXTRACK_IDE_CMD" ]; then

  echo "Using the project IDE cmd: $IDE"

  IDE="$HELIXTRACK_IDE_CMD"

else 

  echo "Using the default project IDE cmd: $IDE"
fi

# TODO: Move under the Software-Toolkit responsibility and make it reusable; Incorporate it into the iconify script
# TODO: Change the SUBMODULES_HOME variable into: PROJECT_TOOLKIT_HOME; Leave SUBMODULES_HOME where it makes sense - work with submodules.
#
RUN_IN_TERMINAL() {

  if [ -z "$1" ]; then

    echo "ERROR: Command to run parameter is mandatory"
    exit 1
  fi

  COMMAND_TO_RUN="$1"

  if which mate-terminal >/dev/null 2>&1; then

    mate-terminal --geometry=250x70 -e "$COMMAND_TO_RUN"
    
  else

    gnome-terminal --geometry=250x70 -- /bin/bash -ic "source ~/.bashrc && $COMMAND_TO_RUN; read"
  fi
}

OPEN_PROJECT="sh $SCRIPT_OPEN $IDE $PROJECT"
OPEN_TOOLKIT="sh $SCRIPT_OPEN $IDE $SUBMODULES_HOME"

RUN_IN_TERMINAL "$OPEN_PROJECT" && RUN_IN_TERMINAL "$OPEN_TOOLKIT"
