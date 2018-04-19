FROM golang:1.8
ADD . /go/src/github.com/ShepBook/k8s-solr
WORKDIR /go/src/github.com/ShepBook/k8s-solr
RUN go get github.com/mitchellh/gox
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/src/github.com/ShepBook/k8s-solr/bin/k8s-solr_linux_amd64 /usr/local/bin/k8s-solr

CMD ["k8s-solr"]
