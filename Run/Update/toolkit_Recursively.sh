#!/bin/bash

HERE="$(pwd)"

cd "$HERE/Core" && sh Toolkit/update_software_toolkit_recursively.sh "$HERE"