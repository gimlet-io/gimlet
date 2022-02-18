#!/bin/bash      

cat << EOF >> $PWD/.env

# Gimlet Dashboard

HOST=$(gp url 9000)
JWT_SECRET=$(openssl rand -hex 32)
GIMLETD_URL=
GIMLETD_TOKEN=
GITHUB_APP_ID=$GITHUB_APP_ID
GITHUB_INSTALLATION_ID=$GITHUB_INSTALLATION_ID
GITHUB_PRIVATE_KEY=$(echo $GITHUB_PRIVATE_KEY | base64 -d)
GITHUB_CLIENT_ID=$GITHUB_CLIENT_ID
GITHUB_CLIENT_SECRET=$GITHUB_CLIENT_SECRET
GITHUB_DEBUG=true
GITHUB_ORG=$GITHUB_ORG
REPO_CACHE_PATH=/workspace/gimlet/repocache
EOF

# Replace proxy url
sed -i -e "s|<<REPLACE ME>>|$(gp url 9000)|" web/dashboard/package.json

# Known hosts for SSH
sudo cp docker/dashboard/known_hosts /etc/ssh/ssh_known_hosts
