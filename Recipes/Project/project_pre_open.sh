#!/bin/bash

HERE=$(pwd)

SCRIPT_CORE_RECIPE_PRE_OPEN="$HERE/Core/Recipes/Project/project_pre_open.sh"

# TODO: Reconnect
#
# SCRIPT_PROPRIATERY_RECIPE_PRE_OPEN="$HERE/Propriatery/Recipes/Project/project_pre_open.sh"

EXECUTE_RECIPE() {

    SCRIPT="$1"

    if test -e "$SCRIPT"; then

        if sh "$SCRIPT" >/dev/null 2>&1; then

            echo "Recipe executed with success: '$SCRIPT'"

        else

            echo "ERROR: Recipe failed, '$SCRIPT'"
            exit 1
        fi

    else

        echo "WARNING: No recipe found at $SCRIPT"
    fi
}

if [ -z "$GENERAL_SERVER" ]; then

    echo "ERROR: 'GENERAL_SERVER' variable is not set"
    exit 1
fi

echo "Checking the server availability: '$GENERAL_SERVER'"

if ! ping -c 2 "$GENERAL_SERVER"; then

    echo "ERROR: '$GENERAL_SERVER' is not accessible"
    exit 1
fi

EXECUTE_RECIPE "$SCRIPT_CORE_RECIPE_PRE_OPEN"

# FIXME: Re-enable when the Propriatery module is reconnected
#
# EXECUTE_RECIPE "$SCRIPT_PROPRIATERY_RECIPE_PRE_OPEN"
