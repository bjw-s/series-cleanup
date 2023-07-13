FROM docker.io/library/golang:1.20-alpine@sha256:7839c9f01b5502d7cb5198b2c032857023424470b3e31ae46a8261ffca72912a as build
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

FROM gcr.io/distroless/static:nonroot@sha256:9ecc53c269509f63c69a266168e4a687c7eb8c0cfd753bd8bfcaa4f58a90876f
USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /build/series-cleanup /app/series-cleanup
COPY --from=build --chown=nonroot:nonroot /sbin/tini-static /sbin/tini
ENTRYPOINT [ "/sbin/tini", "--", "/app/series-cleanup" ]
VOLUME "/config"
LABEL \
  org.opencontainers.image.base.name="gcr.io/distroless/static:nonroot" \
  org.opencontainers.image.title="series-cleanup" \
  org.opencontainers.image.source="https://github.com/bjw-s/series-cleanup"
