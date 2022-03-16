#!/usr/bin/env bash

HOST=$1

if [ -z "$HOST" ]
  then
    echo "usage:"
    echo "  curl -s https://get.gimlet.io | bash -s <<gimlet.your-domain.com>>"
    exit -1
fi

echo "TODO print context name"
echo "TODO print namespace name"
echo "TODO ask for confirmation"

kubectl run gimlet-installer --image=ghcr.io/gimlet-io/installer:latest

echo "TODO print host file edit script"
echo "TODO then visit $HOST"
