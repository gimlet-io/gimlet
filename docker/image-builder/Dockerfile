FROM alpine:3.17.1

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

RUN addgroup -S image-builder && adduser -S image-builder -G image-builder

COPY --chown=image-builder:image-builder bin/image-builder-linux-x86_64 /bin/image-builder

USER image-builder

EXPOSE 5000
CMD ["/bin/image-builder"]
