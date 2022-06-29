#!/usr/bin/env bash

HOST=$1
ORG=$2
VERSION="v0.3.11"

if [ -z "$HOST" ]
  then
    echo "usage:"
    echo "  curl -s https://get.gimlet.io | bash -s <<your-domain.com>> [<<your-github-org>>]"
    exit -1
fi

echo ""
echo "⏳ Downloading Gimlet installer.."
curl -L https://github.com/gimlet-io/gimlet/releases/download/installer-$VERSION/gimlet-installer-$(uname)-$(uname -m) -o gimlet-installer
chmod +x gimlet-installer

echo ""
echo "👉 Point $HOST to localhost temporarily with:"
echo "sudo sh -c 'echo 127.0.0.1 gimlet.$HOST >> /etc/hosts'"
echo ""

read -p "Press 'y' when you are ready with the host file change " REPLY < /dev/tty
echo ""
if ! [[ $REPLY =~ ^[Yy]$ ]]
then
  echo "Aborted"
  exit 1
fi

echo ""
echo "⏳ Starting Gimlet installer.."
echo "We are going to need sudo to run it on port 443"

echo ""
echo "👉 Once started, open https://gimlet.$HOST and follow the installer steps"

sudo HOST=$HOST ORG=$ORG ./gimlet-installer

echo ""
echo "👉 Remove the host file entry now with"
echo "sudo nano /etc/hosts"
echo ""

kubectl get svc -n infrastructure

echo ""
echo "👉 Add gimlet.$HOST to your DNS server"
echo "Point it to the External IP of the ingress-nginx service"
echo "kubectl get svc -n infrastructure"
echo ""
echo "👉 Visit https://gimlet.$HOST"
echo ""


