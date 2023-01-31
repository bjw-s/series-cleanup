FROM gcr.io/distroless/static:nonroot@sha256:26d07ba1f954c02943786e352bc2c8f4eac719ae2f76a0ced68a953bed93a779
COPY series-cleanup /usr/local/bin/series-cleanup
USER nonroot:nonroot
VOLUME [ "/data" ]
ENTRYPOINT ["series-cleanup"]
CMD ["-c", "/data/settings.json"]
