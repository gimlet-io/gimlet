FROM alpine:3.16.0

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

ENV DATABASE_DRIVER=sqlite3
ENV DATABASE_CONFIG=/var/lib/gimletd/gimletd.sqlite
ENV XDG_CACHE_HOME /var/lib/gimletd

ADD bin/gimletd-linux-x86_64 /bin/gimletd

RUN addgroup -S gimletd && adduser -S gimletd -G gimletd

ADD docker/gimletd/known_hosts /etc/ssh/ssh_known_hosts

RUN mkdir /var/lib/gimletd
RUN chown gimletd:gimletd /var/lib/gimletd

USER gimletd
WORKDIR /var/lib/gimletd

RUN git config --global user.name "GimletD"
RUN git config --global user.email "gimletd@gimlet.io"

EXPOSE 8888
CMD ["/bin/gimletd"]
