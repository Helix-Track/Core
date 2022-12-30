#!/bin/bash

# TODO: Don't rebuild if Build directory already exists
cd ../Core/Core &&
sh Versionable/versionable_build.sh Application .. && Application/Build/htCore --help