# Gimlet

[![Go Report Card](https://goreportcard.com/badge/github.com/gimlet-io/gimlet)](https://goreportcard.com/report/github.com/gimlet-io/gimlet)

Gimlet is a deployment tool built on Kubernetes to make the deploy, preview and rollback workflows accessible to everyone.

- [Concepts](https://gimlet.io/docs/concepts)
- [Documentation](https://gimlet.io/docs/)

## Installation

### Prerequisites

- A [Github.com](https://github.com) personal or organization account.
- A Kubernetes cluster running on your laptop or on a cloud provider. We recommend using k3d on your laptop if you are evaluating Gimlet. It takes only a single command to start one, and it runs in a container.

### Optional: Launching k3d on your laptop

K3d is a lightweight Kubernetes cluster that runs in a container on your laptop. At Gimlet, we use k3d solely for our local needs and we recommend you do the same.

Install k3d with:

```
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

Then launch a cluster:

```
k3d cluster create gimlet-cluster --k3s-arg "--disable=traefik@server:0"
```

Once your cluster is up, validate it with `kubectl get nodes`:

```
INFO[0000] Prep: Network
INFO[0000] Created network 'k3d-gimlet-cluster'
INFO[0000] Created image volume k3d-gimlet-cluster-images
INFO[0000] Starting new tools node...
INFO[0000] Starting Node 'k3d-gimlet-cluster-tools'
INFO[0001] Creating node 'k3d-gimlet-cluster-server-0'
INFO[0001] Creating LoadBalancer 'k3d-gimlet-cluster-serverlb'
INFO[0001] Using the k3d-tools node to gather environment information
INFO[0001] Starting new tools node...
INFO[0001] Starting Node 'k3d-gimlet-cluster-tools'
INFO[0002] Starting cluster 'my-first-cluster'
INFO[0002] Starting servers...
INFO[0003] Starting Node 'k3d-gimlet-cluster-server-0'
INFO[0009] All agents already running.
INFO[0009] Starting helpers...
INFO[0009] Starting Node 'k3d-gimlet-cluster-serverlb'
INFO[0016] Injecting records for hostAliases (incl. host.k3d.internal) and for 3 network members into CoreDNS configmap...
INFO[0018] Cluster 'my-first-cluster' created successfully!
INFO[0018] You can now use it like this:
kubectl cluster-info

$ kubectl get nodes
NAME                          STATUS   ROLES                  AGE   VERSION
k3d-gimlet-cluster-server-0   Ready    control-plane,master   11s   v1.26.4+k3s1
```

## Install Gimlet with a oneliner

```
kubectl apply -f https://raw.githubusercontent.com/gimlet-io/gimlet/main/deploy/gimlet.yaml
```

Then access it with port-forward on [http://127.0.0.1:9000](http://127.0.0.1:9000)

```
kubectl port-forward svc/gimlet 9000:9000
```

### Admin password

You can find the admin password in the logs:

```
$ kubectl logs deploy/gimlet | grep "Admin auth key"
time="2023-07-14T14:28:59Z" level=info msg="Admin auth key: 1c04722af2e830c319e590xxxxxxxx" file="[dashboard.go:55]"
```

### Alternative installation method

We generate the Kubernetes manifests from a Helm chart. You can use this configuration directly with Helm if you prefer.

```
helm template gimlet onechart/onechart \
  -f https://raw.githubusercontent.com/gimlet-io/gimlet/main/fixtures/gimlet-helm-values.yaml
```

For all Gimlet environment variables, see the [Gimlet configuration reference](https://gimlet.io/docs/reference/gimlet-configuration-reference).

## License

We switched to a source-available license with source code still hosted on Github.

[Pricing](https://gimlet.io/pricing)

#### Self-host

If you self-host Gimlet for commercial purposes, you need to purchase the license, which is $300 for a year. 
Non-profit and individual use of Gimlet is free and comes without usage limitations.

#### Cloud

You'll start with a 7-day trial to evaluate if Gimlet is helpful for you.

On the technical level, you'll start off with an ephemeral infrastructure provided by us for the trial period and can't connect any real clusters until you purchase the license.

After your purchase, you can connect your own Kubernetes clusters and move your applications to this permanent infrastructure.

Your license will expire a year after your purchase.

#### FAQ: I self-hosted Gimlet before the license change. How can I use it now?

We switched to a source-available license with source code still hosted on Github.

You are required to purchase a license once you upgrade to Gimlet 1.0 and above. As per our pricing you may be eiligible to our free tier.

You are also free to write to us and we probably grant a free license to you. We are grateful for people who trusted us from the early days.

## Contact Us

You can contact us in the issues in this repository, or you can reach us on our Discord server at the link below:
- [Gimlet Discord](https://discord.com/invite/ZwQDxPkYzE)
