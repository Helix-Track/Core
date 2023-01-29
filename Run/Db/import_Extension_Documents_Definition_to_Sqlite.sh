#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && sh Run/Include/Database/import_to_sqlite.sh Core/Database/DDL/Extensions/Documents/Definition.V1.sql