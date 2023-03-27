FROM docker.io/library/golang:1.20-alpine@sha256:576da1aa73f8ffa3dc6a7577b6032bd834aa84f2e1714d3e7e96b06b49f4e177 as build
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

FROM gcr.io/distroless/static:nonroot@sha256:ddde70b96543be368208791ad8ddc9b483cbb33a67bce861e73ad519bc0ed616
USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /build/series-cleanup /app/series-cleanup
COPY --from=build --chown=nonroot:nonroot /sbin/tini-static /sbin/tini
ENTRYPOINT [ "/sbin/tini", "--", "/app/series-cleanup" ]
VOLUME "/config"
LABEL \
  org.opencontainers.image.base.name="gcr.io/distroless/static:nonroot" \
  org.opencontainers.image.title="series-cleanup" \
  org.opencontainers.image.source="https://github.com/bjw-s/series-cleanup"
