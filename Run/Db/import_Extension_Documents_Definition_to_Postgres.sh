#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && sh Run/Include/Database/import_to_postgres.sh Database/DDL/Extensions/Documents/Definition.V1.sql