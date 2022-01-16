FROM gcr.io/distroless/static:nonroot
COPY series-cleanup /usr/local/bin/series-cleanup
USER nonroot:nonroot
VOLUME [ "/data" ]
ENTRYPOINT ["series-cleanup"]
CMD ["-c", "/data/settings.json"]
