FROM --platform=$TARGETPLATFORM ubuntu:jammy

ARG cnb_uid=1000
ARG cnb_gid=1000

# Install packages that we want to make available at both build and run time
RUN apt update && \
  apt install -y xz-utils ca-certificates libyaml-0-2 && \
  rm -rf /var/lib/apt/lists/*

# Create user and group
RUN groupadd cnb --gid ${cnb_gid} && \
  useradd --uid ${cnb_uid} --gid ${cnb_gid} -m -s /bin/bash cnb

# Set required CNB information
ENV CNB_USER_ID=${cnb_uid}
ENV CNB_GROUP_ID=${cnb_gid}