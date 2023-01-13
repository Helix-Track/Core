#!/bin/bash

cd "$(dirname "$0")"

sh Database/import_to_sqlite.sh ../Core/Database/DDL/Definition.V1.sql