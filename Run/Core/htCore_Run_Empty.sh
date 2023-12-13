#!/bin/bash

HERE="$(pwd)"

cd "$HERE/Core" &&
    bash Versionable/versionable_build.sh Application .. true && Application/Build/htCore -l -d -c Configurations/empty.json