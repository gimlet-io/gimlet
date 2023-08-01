FROM gcr.io/buildpacks/builder:v1

COPY --chown=image-builder:image-builder bin/image-builder-linux-x86_64 /bin/image-builder

ENV CNB_PLATFORM_API=0.7

EXPOSE 5000
CMD ["/bin/image-builder"]
