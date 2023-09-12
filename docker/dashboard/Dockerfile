FROM --platform=linux/arm64 alpine:3.18.3

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

ENV DATABASE_DRIVER=sqlite
ENV DATABASE_CONFIG=/var/lib/gimlet-dashboard/gimlet-dashboard.sqlite?_pragma=busy_timeout=10000
ENV XDG_CACHE_HOME /var/lib/gimlet-dashboard
ENV GIT_ROOT=/var/lib/gimlet-dashboard/git-root/

RUN addgroup -S gimlet-dashboard && adduser -S gimlet-dashboard -G gimlet-dashboard

ADD docker/dashboard/known_hosts /etc/ssh/ssh_known_hosts

RUN mkdir /var/lib/gimlet-dashboard
RUN chown gimlet-dashboard:gimlet-dashboard /var/lib/gimlet-dashboard
WORKDIR /gimlet-dashboard

ARG TARGETPLATFORM
ARG BUILDPLATFORM

# COPY --chown=gimlet-dashboard:gimlet-dashboard bin/${TARGETPLATFORM}/gimlet-dashboard gimlet-dashboard
# COPY --chown=gimlet-dashboard:gimlet-dashboard web/dashboard/build ./web/build/

USER gimlet-dashboard

RUN git config --global user.name "Gimlet"
RUN git config --global user.email "gimlet@gimlet.io"
RUN git config --global init.defaultBranch main

EXPOSE 9000
CMD ["/gimlet-dashboard/gimlet-dashboard"]
