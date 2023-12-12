#!/bin/bash

if [ -z "$SUBMODULES_HOME" ]; then

  echo "ERROR: SUBMODULES_HOME not available"
  exit 1
fi

SCRIPT_GET_HTTPD="$SUBMODULES_HOME/Software-Toolkit/Utils/Httpd/get_httpd.sh"

if ! test -e "$SCRIPT_GET_HTTPD"; then

    echo "ERROR: Script not found '$SCRIPT_GET_HTTPD'"
    exit 1
fi

if [ -z "$HTTPD_WEB_ROOT_SHARED" ]; then

    echo "ERROR: Env. variable not defined 'HTTPD_WEB_ROOT_SHARED'"
    exit 1
fi

sh "$SCRIPT_GET_HTTPD" "httpd.$(hostname)" "$HTTPD_WEB_ROOT_SHARED" 8081