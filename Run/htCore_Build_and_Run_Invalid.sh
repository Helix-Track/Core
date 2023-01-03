#!/bin/bash

cd ../Core/Core &&
sh Versionable/versionable_build.sh Application .. && Application/Build/htCore -l -d -c Configurations/invalid.json