#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && 
    bash Run/Db/import_All_Definitions_to_Sqlite.sh && \
    bash Run/Db/import_All_Definitions_to_Postgres.sh && \
    bash Run/Prepare/core.sh && bash Run/Prepare/propriatery.sh
