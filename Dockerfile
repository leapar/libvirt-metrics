FROM alpine

RUN set -ex && \
    apk update && \
    apk add ca-certificates g++ git go wget libxml2  libxml2-dev  libvirt-dev && \
    mkdir -p /go/src /go/bin && chmod -R 777 /go && \
    export GOPATH=/go && go get github.com/leapar/libvirt-metrics && strip /go/bin/libvirt-metrics && \
    apk del g++ git go wget libxml2-dev

ADD libvirt-metrics.json /etc/libvirt-metrics.json
ENV PATH /go/bin:$PATH

RUN echo "Asia/Shanghai" > /etc/timezone

CMD ["libvirt-metrics"]