#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && bash Run/Include/Database/import_to_sqlite.sh Database/DDL/Extensions/Documents/Definition.V1.sql