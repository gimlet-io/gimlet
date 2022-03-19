#!/usr/bin/env bash

HOST=$1
VERSION="v0.2.0"

if [ -z "$HOST" ]
  then
    echo "usage:"
    echo "  curl -s https://get.gimlet.io | bash -s <<gimlet.your-domain.com>>"
    exit -1
fi

echo ""
echo "⏳ Starting Gimlet installer pod.."
kubectl run gimlet-installer --image=ghcr.io/gimlet-io/installer:$VERSION

echo ""
echo "👉 Point $HOST to localhost temporarily with:"
echo "sudo echo "127.0.0.1 $HOST" >> /etc/hosts"
echo ""
echo "👉 Forward the installer to $HOST with:"
echo "sudo KUBECONFIG=$HOME/.kube/config kubectl port-forward pod/gimlet-installer 443:4443"
echo ""
echo "👉 visit $HOST to access the installer"
echo ""
