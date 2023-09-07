#!/bin/bash

HERE="$(pwd)"

SCRIPT_CORE_RECIPE_PRE_OPEN="$HERE/Core/Recipes/project_pre_open.sh"
SCRIPT_PROPRIATERY_RECIPE_PRE_OPEN="$HERE/Propriatery/Recipes/project_pre_open.sh"

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

    DIR_SUBMODULES="_Submodules"
    DIR_SOURCE="$HERE/$DIR_SUBMODULES/$MODULE_SOURCE"
    DIR_DESTINATION="$HERE/$MODULE_DESTINATION"

    if test -e "$DIR_SOURCE"; then

        echo "Linking: $DIR_SOURCE -> $DIR_DESTINATION"

        if test -e "$DIR_DESTINATION"; then

            if rm -f "$DIR_DESTINATION"; then

                echo "Link removed"

            else

                echo "ERROR: Link failed to remove '$DIR_DESTINATION'"
                exit 1
            fi
        fi

        if ln -s "$DIR_SOURCE" "$DIR_DESTINATION" && test -e "$DIR_DESTINATION"; then

            echo "Linking success"

        else

            echo "ERROR: Could not create symbolic link '$DIR_DESTINATION'"
        fi

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