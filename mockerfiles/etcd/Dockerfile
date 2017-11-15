FROM alpine:latest

RUN apk --no-cache add ca-certificates

ADD https://github.com/coreos/etcd/releases/download/v3.2.8/etcd-v3.2.8-linux-amd64.tar.gz etcd-v3.2.8-linux-amd64.tar.gz
RUN tar -xvf etcd-v3.2.8-linux-amd64.tar.gz && \
    mv etcd-v3.2.8-linux-amd64/etcd* /usr/local/bin/

CMD ["/usr/local/bin/etcd", "--force-new-cluster"]