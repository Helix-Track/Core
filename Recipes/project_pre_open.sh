#!/bin/bash

HERE="$(pwd)"

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

LINK_MODULE() {

    if [ -z "$1" ]; then

        ecgo "ERROR: Module source parameter is mandatory"
        exit 1
    fi

    if [ -z "$2" ]; then

        ecgo "ERROR: Module destination parameter is mandatory"
        exit 1
    fi

    PREFIX=""

    if [ -n "$3" ]; then

        PREFIX="$3"
    fi

    HERE="$(pwd)"
    MODULE_SOURCE="$1"
    MODULE_DESTINATION="${PREFIX}$2"

    DIR_SUBMODULES="_SUBMODULES"
    DIR_SOURCE="$HERE/$DIR_SUBMODULES/$MODULE_SOURCE"
    DIR_DESTINATION="$HERE/$MODULE_DESTINATION"

    if test -e "$DIR_SOURCE"; then

        # TODO: Implement linking

        echo "Linking: $DIR_SOURCE -> $DIR_DESTINATION"

    else

        echo "ERROR: Source linking directory does not exist '$DIR_SOURCE'"
        exit 1
    fi
}

PREFIX="module_"

LINK_MODULE "Upstreamable" "Upstreamable" "$PREFIX"
LINK_MODULE "Software-Toolkit" "Toolkit" "$PREFIX"

EXECUTE_RECIPE "$SCRIPT_CORE_RECIPE_PRE_OPEN"
EXECUTE_RECIPE "$SCRIPT_PROPRIATERY_RECIPE_PRE_OPEN"