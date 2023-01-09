FROM alpine:3.17.1

RUN apk --no-cache add bash curl git openssl jq

WORKDIR /action

ADD bin/gimlet-linux-x86_64 /bin/gimlet
