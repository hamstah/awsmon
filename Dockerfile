FROM golang:alpine as builder

ADD ./vendor /go/src/github.com/cirocosta/go-aws-mon/vendor
ADD ./awsmon /go/src/github.com/cirocosta/go-aws-mon/awsmon

WORKDIR /go/src/github.com/cirocosta/go-aws-mon/awsmon
RUN set -ex && \
  CGO_ENABLED=0 go build -v -a -ldflags '-extldflags "-static"' && \
  mv ./awsmon /usr/bin/awsmon

FROM busybox
COPY --from=builder /usr/bin/awsmon /awsmon

CMD [ "awsmon" ]
