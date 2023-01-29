#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && sh Run/Include/Database/import_to_postgres.sh Core/Database/Test/Test.Postgres.sql