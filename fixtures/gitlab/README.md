# Testing Gitlab

## Starting Gitlab
```
docker-compose up -d
docker network inspect gitlab_gitlab | grep IPv4Address
docker-compose exec -it web cat /etc/gitlab/initial_root_password
sudo sh -c 'echo 172.26.0.2 gitlab.local >> /etc/hosts'
```

## Accessing it from k3d

```
k3d cluster create gitlab3 --network gitlab_gitlab --host-alias 172.26.0.2:gitlab.local
```

Gimletd known hosts file in this case
```
    vars:
      SSH_KNOWN_HOSTS: /var/lib/gimletd/gitlab_hosts/gitlab.key:/etc/ssh/ssh_known_hosts
```

```
    fileSecrets:
      - name:  gitlab-hosts
        path: /var/lib/gimletd/gitlab_hosts
        secrets:
          gitlab.key: |
            gitlab.local ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlz... same value as in the flux/deploy key file
```

in local `.env` file:

```
SSH_KNOWN_HOSTS=/home/laszlo/projects/gimlet/fixtures/gitlab/gitlab.local.hosts:/home/laszlo/.ssh/known_hosts
```
