#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && 
    sh Run/Db/import_All_Definitions_to_Sqlite.sh && \
    sh Run/Db/import_All_Definitions_to_Postgres.sh && \
    sh Run/Prepare/core.sh && sh Run/Prepare/propriatery.sh
