#!/bin/bash

cd ../Core/Core &&
sh Versionable/versionable_build.sh Application .. && Application/Build/htCore -l -c Configurations/dev.json