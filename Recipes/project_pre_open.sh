#!/bin/bash

HERE="$(pwd)"
PROJECT="$HERE"

SCRIPT_CORE_RECIPE_PRE_OPEN="$HERE/Core/Recipes/project_pre_open.sh"
SCRIPT_PROPRIATERY_RECIPE_PRE_OPEN="$HERE/Propriatery/Recipes/project_pre_open.sh"

EXECUTE_RECIPE() {

    SCRIPT="$1"

    if test -e "$SCRIPT"; then

        if ! sh "$SCRIPT"; then

            echo "ERROR: Recipe failed, $SCRIPT"
            exit 1
        fi
    else

        echo "WARNING: No recipe found at $SCRIPT"
    fi
}

EXECUTE_RECIPE "$SCRIPT_CORE_RECIPE_PRE_OPEN"
EXECUTE_RECIPE "$SCRIPT_PROPRIATERY_RECIPE_PRE_OPEN"