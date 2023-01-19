#!/bin/bash

cd "$(dirname "$0")"

sh Database/import_to_postgres.sh ../Core/Database/Test/Test.Postgres.sql