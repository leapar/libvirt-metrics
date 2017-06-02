FROM alpine

RUN set -ex && \
    apk update && \
    apk add ca-certificates g++ git go libnl-dev linux-headers make perl pkgconf libtirpc-dev wget openssl && \
    update-ca-certificates