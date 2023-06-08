```
tar -cvzf gimlet-documentation.tar.gz -C gimlet-documentation-bare .

make build-image-builder
docker run --rm -it -v "$PWD":/usr/src/app -w /usr/src/app -p 9000:9000 paketobuildpacks/builder ./build/image-builder

curl -F 'data=@/home/laszlo/projects/gimlet-documentation.tar.gz' http://localhost:9000/build-image
curl -F 'image=192.168.50.129:5000/gimlet-documentation:latest' -F 'data=@/home/laszlo/projects/gimlet-documentation.tar.gz' http://localhost:9000/build-image

docker run -it --rm -p 5000:5000 --name registry registry:2

```
containerPort: 5000
image:
  repository: registry
  tag: "2"

```
```


```
helm template registry onechart/onechart -f values.yaml | kubectl apply -f -
helm template app-from-registry onechart/onechart -f values2.yaml | kubectl apply -f -
kubectl port-forward svc/registry 5000:5000 --address=192.168.50.129
curl -F 'image=192.168.50.129:5000/gimlet-documentation:1' -F 'data=@/home/laszlo/projects/gimlet-documentation.tar.gz' http://localhost:9000/build-image
```
