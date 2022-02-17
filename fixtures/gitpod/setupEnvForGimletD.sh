#!/bin/bash      

sudo cp docker/gimletd/known_hosts /etc/ssh/ssh_known_hosts

curl -L https://github.com/gimlet-io/gimlet-cli/releases/download/v0.12.2/gimlet-$(uname)-$(uname -m) -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version

mkdir -p ~/.gimlet
cat << EOF > ~/.gimlet/config
export GIMLET_SERVER=http://127.0.0.1:8888
EOF

cat << EOF >> $PWD/.env

# GimletD

GITOPS_REPO=$GITOPS_REPO
GITOPS_REPO_DEPLOY_KEY_PATH=/workspace/gimlet/deploykey
PRINT_ADMIN_TOKEN=true

NOTIFICATIONS_PROVIDER=slack
NOTIFICATIONS_TOKEN=$SLACK_TOKEN
NOTIFICATIONS_DEFAULT_CHANNEL=gimletd
EOF

echo $DEPLOY_KEY | base64 -d > deploykey
echo $DEPLOY_KEY_PUB | base64 -d > deploykey.pub
chmod 500 deploykey
chmod 500 deploykey.pub
