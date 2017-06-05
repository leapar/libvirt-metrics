FROM alpine

RUN set -ex && \
    apk update && \
    apk add ca-certificates g++ git go libnl-dev linux-headers make perl pkgconf libtirpc-dev wget openssl && \
    update-ca-certificates && \
    cd /tmp && \
    wget ftp://xmlsoft.org/libxml2/libxml2-2.9.4.tar.gz && \
    tar -xf libxml2-2.9.4.tar.gz && \
    cd libxml2-2.9.4 && \
    ./configure --disable-shared --enable-static && \
    make -j2 && \
    make install && \
    cd /tmp && \
    wget https://libvirt.org/sources/libvirt-3.2.0.tar.xz && \
    tar -xf libvirt-3.2.0.tar.xz && \
    cd libvirt-3.2.0 && \
    ./configure --disable-shared --enable-static --localstatedir=/var --without-storage-mpath && \
    make -j2 && \
    make install && \
    sed -i 's/^Libs:.*/& -lnl -ltirpc -lxml2/' /usr/local/lib/pkgconfig/libvirt.pc && \
    mkdir /libvirt-metrics && \
    cd /libvirt-metrics && \
    git clone https://github.com/leapar/libvirt-metrics && \