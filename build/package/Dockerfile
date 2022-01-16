FROM alpine:3.11.5 as build
RUN \
   apk --no-cache add ca-certificates

FROM gcr.io/distroless/static:nonroot
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY series-cleanup /usr/local/bin/series-cleanup
USER nonroot:nonroot
VOLUME [ "/data" ]
ENTRYPOINT ["series-cleanup"]
CMD ["-c", "/data/settings.json"]
