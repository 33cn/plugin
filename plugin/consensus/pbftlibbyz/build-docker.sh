#!/bin/sh

cd ../../../
docker build -t pbftlibbyz -f plugin/consensus/pbftlibbyz/Dockerfile .
