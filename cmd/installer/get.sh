#!/usr/bin/env bash

HOST=$1
VERSION="v0.4.7"

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

echo ""
echo "👉 Once started, open http://gimlet.$HOST:9000 and follow the installer steps"

HOST=$HOST ./gimlet-installer

echo " ✅ Installer stopped"

until [ $(kubectl get kustomizations.kustomize.toolkit.fluxcd.io -A | grep gitops-repo | grep True | wc -l) -eq 4 ]
do
  echo ""
  echo " 🧐 Waiting for all four gitops kustomizations become ready, ctrl+c to abort"
  echo ""
  echo "$ kubectl get kustomizations.kustomize.toolkit.fluxcd.io -A"
  kubectl get kustomizations.kustomize.toolkit.fluxcd.io -A | grep -w 'gitops-repo\|READY'
  sleep 3
done

echo ""
echo " ✅ Gitops loop is healthy"
echo ""

until [ $(kubectl get pods -n infrastructure | grep gimlet | grep 1/1 | wc -l) -eq 3 ]
do
  echo ""
  echo " 🧐 Waiting for Gimlet to start up in the cluster, ctrl+c to abort"
  echo ""
  echo "$ kubectl get pods -n infrastructure | grep gimlet"
  kubectl get pods -n infrastructure | grep 'gimlet\|READY\|postgres'
  sleep 3
done

echo ""
echo " ✅ Gimlet is up"
echo ""

kubectl get svc -n infrastructure
kubectl get svc -n kube-system

