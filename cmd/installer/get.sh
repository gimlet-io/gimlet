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
echo "⏳ Starting Gimlet installer pod.."
kubectl run gimlet-installer --image=ghcr.io/gimlet-io/installer:$VERSION --env="HOST=$HOST"

echo ""
echo "👉 Point $HOST to localhost temporarily with:"
echo "sudo echo "127.0.0.1 gimlet.$HOST" >> /etc/hosts"
echo ""
echo "👉 Forward the installer to gimlet.$HOST with:"
echo "sudo KUBECONFIG=$HOME/.kube/config kubectl port-forward pod/gimlet-installer 443:4443"
echo ""
echo "👉 visit https://gimlet.$HOST to access the installer"
echo ""

echo "👉 Once done, remove the host file entry"
echo ""
echo "👉 And add to your DNS the LB IP: kubectl get svc -n infrastructure"
echo ""
