#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && sh Run/Db/import_Main_Definition_to_Sqlite.sh && \
    sh Run/Db/import_Services_Authentication_Definition_to_Sqlite.sh && \
    sh Run/Db/import_Extension_Documents_Definition_to_Sqlite.sh && \
    sh Run/Db/import_Extension_Times_Definition_to_Sqlite.sh && \
    sh Run/Db/import_Extension_Chats_Definition_to_Sqlite.sh
