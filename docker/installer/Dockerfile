FROM alpine:3.13

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

ADD bin/gimlet-installer-linux-x86_64 /bin/gimlet-installer

RUN addgroup -S gimlet-installer && adduser -S gimlet-installer -G gimlet-installer

ADD docker/gimlet-installer/known_hosts /etc/ssh/ssh_known_hosts

RUN mkdir /var/lib/gimlet-installer
RUN chown gimlet-installer:gimlet-installer /var/lib/gimlet-installer

USER gimlet-installer
WORKDIR /var/lib/gimlet-installer

RUN git config --global user.name "gimlet-installer"
RUN git config --global user.email "gimlet-installer@gimlet.io"

EXPOSE 8888
CMD ["/bin/gimlet-installer"]
