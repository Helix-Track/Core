#!/bin/bash

cd "$(dirname "$0")"

sh Database/import_to_sqlite.sh ../Core/Database/DDL/Extensions/Times/Definition.V1.sql