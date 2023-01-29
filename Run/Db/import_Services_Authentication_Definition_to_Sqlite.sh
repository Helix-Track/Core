#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && sh Run/Include/Database/import_to_sqlite.sh Core/Database/DDL/Services/Authentication/Definition.V1.sql