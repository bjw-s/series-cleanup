FROM docker.io/library/golang:1.21-alpine@sha256:7c8aba231ef64956d0b1fb550b358e7a9c5101508d758283f88c10ff90389cbc as build
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}
# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates tini-static \
    && update-ca-certificates
WORKDIR /build
COPY . .
RUN go build -ldflags="-s -w" -o series-cleanup /build/cmd/series-cleanup/.

FROM gcr.io/distroless/static:nonroot@sha256:112a87f19e83c83711cc81ce8ed0b4d79acd65789682a6a272df57c4a0858534
USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /build/series-cleanup /app/series-cleanup
COPY --from=build --chown=nonroot:nonroot /sbin/tini-static /sbin/tini
ENTRYPOINT [ "/sbin/tini", "--", "/app/series-cleanup" ]
VOLUME "/config"
LABEL \
  org.opencontainers.image.base.name="gcr.io/distroless/static:nonroot" \
  org.opencontainers.image.title="series-cleanup" \
  org.opencontainers.image.source="https://github.com/bjw-s/series-cleanup"
