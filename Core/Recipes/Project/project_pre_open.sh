#!/bin/bash

HERE=$(pwd)

cd "$HERE/Core" && sh Recipes/Installable/prepare.sh && cd "$HERE" && echo "Project pre-open - Core module status: READY"