FROM golang:alpine as builder

ADD ./vendor /go/src/github.com/cirocosta/awsmon/vendor
ADD ./main.go /go/src/github.com/cirocosta/awsmon/main.go
ADD ./lib /go/src/github.com/cirocosta/awsmon/lib

WORKDIR /go/src/github.com/cirocosta/awsmon
RUN set -ex && \
  CGO_ENABLED=0 go build -v -a -ldflags '-extldflags "-static"' && \
  mv ./awsmon /usr/bin/awsmon

FROM busybox
COPY --from=builder /usr/bin/awsmon /usr/local/bin/awsmon

CMD [ "awsmon" ]
