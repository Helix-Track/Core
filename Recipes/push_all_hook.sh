#!/bin/bash

# TODO: Re-enable hook and complete the implementation

exit 0

# if [ -z "$1" ]; then

#     echo "ERROR: Upstreams directory path is mandatory"
# fi

# DIR_UPSTREAMS="$1"
# DIR_ROOT="$DIR_UPSTREAMS/.."
# DIR_GIT="$DIR_ROOT/.git"

# if ! test -e "$DIR_GIT"; then

#     echo "ERROR: Not a Git root directory '$DIR_GIT'"
#     exit 1
# fi

# if [ -z "$SUBMODULES_HOME" ]; then

#   echo "ERROR: The SUBMODULES_HOME is not defined"
#   exit 1
# fi

# SCRIPT_FLAGS="$SUBMODULES_HOME/Software-Toolkit/Utils/Git/gather_submodules_flags.sh"
# SCRIPT_GATHER_SUBMODULES="$SUBMODULES_HOME/Software-Toolkit/Utils/Git/gather_submodules.sh"

# if test -e "$SCRIPT_GATHER_SUBMODULES"; then

#     if test -e "$SCRIPT_FLAGS"; then

#         # shellcheck disable=SC1090
#         . "$SCRIPT_FLAGS"

#     else

#         echo "ERROR: Flags Script not found '$SCRIPT_FLAGS'"
#         exit 1
#     fi

#     F_UPDATE_ONLY="${OPEN}${FLAG_UPDATE_ONLY}=${DIR_ROOT}${CLOSE}"
    
#     FLAGS="FLAGS=[${F_UPDATE_ONLY}]"

#     sh "$SCRIPT_GATHER_SUBMODULES" "$FLAGS"

# else

#   echo "ERROR: Script not found '$SCRIPT_GATHER_SUBMODULES'"
#   exit 1
# fi