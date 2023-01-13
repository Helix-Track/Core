#!/bin/bash

cd "$(dirname "$0")"

sh Database/import_to_sqlite.sh ../Core/Database/DDL/Extensions/Documents/Definition.V1.sql