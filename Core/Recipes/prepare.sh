#!/bin/bash

echo "Preparing the 'HelixTrack Core' for the installation" && \
  git submodule init && git submodule update && \
  echo "The 'HelixTrack Core' is prepared for the installation"