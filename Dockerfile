FROM golang:alpine as builder

ADD ./ /go/src/github.com/cirocosta/awsmon
WORKDIR /go/src/github.com/cirocosta/awsmon

RUN set -ex && \
  CGO_ENABLED=0 go build -v -a -ldflags '-extldflags "-static"' && \
  mv ./awsmon /usr/bin/awsmon

FROM alpine
COPY --from=builder /usr/bin/awsmon /usr/local/bin/awsmon

CMD [ "awsmon" ]
