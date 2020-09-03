ARG BASEIMAGE=alpine:3.12
FROM golang:1.15 as builder
WORKDIR /go/src/series-cleanup
COPY src .
RUN \
  unset GOPATH && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/series-cleanup .

FROM ${BASEIMAGE}

COPY --from=builder /go/bin/series-cleanup /usr/local/bin/series-cleanup

RUN \
  apk --no-cache add ca-certificates && \
  chmod -R +x /usr/local/bin

CMD ["/usr/local/bin/series-cleanup", "-c", "/data/settings.json"]

VOLUME [ "/data" ]