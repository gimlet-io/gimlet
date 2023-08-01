ARG base_image
FROM --platform=$BUILDPLATFORM ${base_image}

# Install packages that we want to make available at build time
RUN apt update && \
  apt install -y git wget jq && \
  rm -rf /var/lib/apt/lists/*

COPY ./yj-linux /usr/local/bin/yj

# Set required CNB information
ARG stack_id
ENV CNB_STACK_ID=${stack_id}
LABEL io.buildpacks.stack.id=${stack_id}
