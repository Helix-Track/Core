#!/bin/bash

HERE="$(pwd)"

cd "$HERE/Core" &&
    bash Versionable/versionable_build.sh Application .. && Application/Build/htCore -l -d -c Configurations/unavailable.json