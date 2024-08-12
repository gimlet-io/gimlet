FROM gcr.io/buildpacks/builder:v1

COPY --chown=cnb:cnb bin/image-builder-linux-x86_64 /bin/image-builder

COPY --chown=cnb:cnb docker/image-builder/gimlet-registry.crt /usr/local/share/ca-certificates/gimlet-registry.crt
USER root
RUN apt update && \
  apt install -y nano && \
  rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
USER cnb

ENV CNB_PLATFORM_API=0.7

EXPOSE 5000
CMD ["/bin/image-builder"]
