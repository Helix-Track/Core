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

        if test -e "$DIR_DESTINATION"; then

            if ! rm -f "$DIR_DESTINATION"; then

                echo "ERROR: Link failed to remove '$DIR_DESTINATION'"
                exit 1
            fi
        fi

    else

        echo "ERROR: Source linking directory does not exist '$DIR_SOURCE'"
        exit 1
    fi
}

PREFIX="module_"

LINK_MODULE "Upstreamable" "Upstreamable" "$PREFIX"
LINK_MODULE "Software-Toolkit" "Toolkit" "$PREFIX"
LINK_MODULE "Dependable" "Dependable" "$PREFIX"
LINK_MODULE "Docker-Definitions" "Definitions_Docker" "$PREFIX"
LINK_MODULE "Software-Definitions" "Definitions_Software" "$PREFIX"
LINK_MODULE "Stack-Definitions" "Definitions_Stack" "$PREFIX"
LINK_MODULE "Installable" "Installable" "$PREFIX"
LINK_MODULE "Project" "Project" "$PREFIX"
LINK_MODULE "Services-Toolkit" "Toolkit_Services" "$PREFIX"
LINK_MODULE "Testable" "Testable" "$PREFIX"
LINK_MODULE "Versionable" "Versionable" "$PREFIX"

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
EXECUTE_RECIPE "$SCRIPT_PROPRIATERY_RECIPE_PRE_OPEN"