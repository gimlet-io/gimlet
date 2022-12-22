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
