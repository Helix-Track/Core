#!/bin/bash

HERE="$(pwd)"
PLUGINS="Application/plugins"
GENERATED="Application/generated"

if test -e "$GENERATED"; then

  if ! sudo rm -rf "$GENERATED"; then

    echo "ERROR: Could not remove the '$GENERATED' directory"
    exit 1
  fi
fi

echo "Generating the code"

if mkdir "$GENERATED" && /usr/local/bin/sql2code-0.0.2/sql2code -i \
            Database/DDL/Definition.V1.sql \
            Database/DDL/Extensions/Chats/Definition.V1.sql \
            Database/DDL/Extensions/Documents/Definition.V1.sql \
            Database/DDL/Extensions/Times/Definition.V1.sql \
            -t cpp -o "$GENERATED"; then

    echo "Code generated"

else

    echo "ERROR: Code not generated"

    exit 1
fi

if test -e "$PLUGINS"; then

  if ! rm -rf "$PLUGINS"; then

    echo "ERROR: Could not remove '$PLUGINS'"
    exit 1
  fi
fi

if ! mkdir -p "$PLUGINS"; then

  echo "ERROR: Could not create '$PLUGINS'"
  exit 1
fi

if [ -z "$DEPENDABLE_DEPENDENCIES_HOME" ]; then

  DEPENDABLE_DEPENDENCIES_HOME="$HERE"
fi

echo "The dependencies home directory: '$DEPENDABLE_DEPENDENCIES_HOME'"

DEPENDENCIES_WORKING_DIRECTORY="$DEPENDABLE_DEPENDENCIES_HOME/_Dependencies"

echo "The dependencies working directory: '$DEPENDENCIES_WORKING_DIRECTORY'"

if cp "$DEPENDENCIES_WORKING_DIRECTORY/Cache/JWT-Drogon/Library/JWT.*" "$PLUGINS" && \
  cp "$DEPENDENCIES_WORKING_DIRECTORY/Cache/JWT-Drogon/Library/JWT*.*" "$PLUGINS"; then

  echo "Drogon JWT plugin copied"
else

  echo "ERROR: Drogon JWT plugin copied not copied"
  exit 1
fi

echo "Preparing the 'HelixTrack Core' for the installation" && \
  git submodule init && git submodule update && \
  echo "The 'HelixTrack Core' is prepared for the installation"