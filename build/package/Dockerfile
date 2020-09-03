FROM golang:1.15-alpine AS build
ENV GO111MODULE=on \
    CGO_ENABLED=0
WORKDIR /build
COPY . .
RUN \
   apk --no-cache add ca-certificates \
   && go build -ldflags "-s -w" -o /go/bin/series-cleanup -v cmd/series-cleanup/main.go \
   && chmod +x /go/bin/series-cleanup

FROM gcr.io/distroless/static:nonroot
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/series-cleanup /usr/local/bin/series-cleanup
USER nonroot:nonroot
VOLUME [ "/data" ]
ENTRYPOINT ["series-cleanup"]
CMD ["-c", "/data/settings.json"]