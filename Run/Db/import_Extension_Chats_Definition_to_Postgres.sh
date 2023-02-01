#!/bin/bash

HERE="$(pwd)"

cd "$HERE" && sh Run/Include/Database/import_to_postgres.sh Core/Database/DDL/Extensions/Chats/Definition.V1.sql