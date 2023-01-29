#!/bin/bash

HERE="$(pwd)"

cd "$HERE/Core" &&
    sh Versionable/versionable_build.sh Application .. true && Application/Build/htCore -l -d -c Configurations/unavailable.json