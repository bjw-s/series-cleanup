FROM docker.io/library/golang:1.21-alpine@sha256:92db786880115648313e75bc7c963d4d37a4a40546ce571ac79de38bd0e4cb29 as build
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

FROM gcr.io/distroless/static:nonroot@sha256:92d40eea0b5307a94f2ebee3e94095e704015fb41e35fc1fcbd1d151cc282222
USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /build/series-cleanup /app/series-cleanup
COPY --from=build --chown=nonroot:nonroot /sbin/tini-static /sbin/tini
ENTRYPOINT [ "/sbin/tini", "--", "/app/series-cleanup" ]
VOLUME "/config"
LABEL \
  org.opencontainers.image.base.name="gcr.io/distroless/static:nonroot" \
  org.opencontainers.image.title="series-cleanup" \
  org.opencontainers.image.source="https://github.com/bjw-s/series-cleanup"
