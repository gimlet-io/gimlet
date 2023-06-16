```
tar -cvzf gimlet-documentation.tar.gz -C gimlet-documentation-bare .

make build-image-builder
docker run --rm -it -v "$PWD":/usr/src/app -w /usr/src/app -p 9000:9000 paketobuildpacks/builder ./build/image-builder

curl -F 'data=@/home/laszlo/projects/gimlet-documentation.tar.gz' http://localhost:9000/build-image

```
