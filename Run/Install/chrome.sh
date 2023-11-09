#!/bin/bash

if epm play chrome; then

    echo "Google Chrome installation completed with success"

else

    echo "Google Chrome installation failed"
    exit 1
fi