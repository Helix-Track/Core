#!/bin/bash

cd "$(dirname "$0")"

sh db_Import_Main_Definition_to_Sqlite.sh && \
    sh db_Import_Services_Authentication_Definition_to_Sqlite.sh && \
    sh db_Import_Extensions_Documents_Definition_to_Sqlite.sh && \
    sh db_Import_Extensions_Times_Definition_to_Sqlite.sh && \
    sh db_Import_Extensions_Chats_Definition_to_Sqlite.sh
