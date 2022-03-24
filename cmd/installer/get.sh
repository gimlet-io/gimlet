#!/usr/bin/env bash

HOST=$1
VERSION="v0.2.5"

if [ -z "$HOST" ]
  then
    echo "usage:"
    echo "  curl -s https://get.gimlet.io | bash -s <<your-domain.com>>"
    exit -1
fi

echo ""
echo "⏳ Downloading Gimlet installer.."
curl -L https://github.com/gimlet-io/gimlet/releases/download/installer-$(version)/gimlet-installer-$(uname)-$(uname -m) -o gimlet-installer
chmod +x gimlet-installer

echo ""
echo "👉 Point $HOST to localhost temporarily with:"
echo "sudo sh -c 'echo 127.0.0.1 gimlet.$HOST >> /etc/hosts'"
echo ""

echo "⏳ Starting Gimlet installer.."
echo "We are going to need sudo to run it on port 443"

echo ""
echo "👉 Once started, open https://gimlet.$HOST and follow the installer steps"
echo ""

echo ""
sudo ./gimlet-installer $HOST

echo ""
echo "👉 Once done, remove the host file entry"
echo ""
echo "👉 And add to your DNS the LB IP: kubectl get svc -n infrastructure"
echo ""
