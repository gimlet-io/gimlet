FROM alpine:3.17.1

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

ADD bin/gimlet-installer-linux-x86_64 /bin/gimlet-installer

RUN addgroup -S gimlet-installer && adduser -S gimlet-installer -G gimlet-installer

USER gimlet-installer

EXPOSE 14000
CMD ["/bin/gimlet-installer"]
