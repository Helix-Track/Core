#!/bin/bash

GENERATED="Application/generated"

if test -e "$GENERATED"; then

  if ! sudo rm -rf "$GENERATED"; then

    echo "ERROR: Could not remove the '$GENERATED' directory"
    exit 1
  fi
fi

echo "Generating the code"

if mkdir "$GENERATED" && /usr/local/bin/sql2code-0.0.1/sql2code -i \
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

echo "Preparing the 'HelixTrack Core' for the installation" && \
  git submodule init && git submodule update && \
  echo "The 'HelixTrack Core' is prepared for the installation"