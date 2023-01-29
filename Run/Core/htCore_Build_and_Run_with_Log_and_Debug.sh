#!/bin/bash

HERE="$(pwd)"

cd "$HERE/Core" &&
    sh Versionable/versionable_build.sh Application .. && Application/Build/htCore -l -d
