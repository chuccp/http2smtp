# syntax=docker/dockerfile:1
FROM busybox:latest

ARG goarch=amd64
ARG release_base=https://github.com/chuccp/http2smtp/releases/latest/download

# Package the GitHub Release binary as-is. Templates, static assets and locale files
# are already embedded by the release build.
ADD ${release_base}/http2smtp-linux-${goarch}.tar.gz /app/
RUN tar -xzf http2smtp-linux-${goarch}.tar.gz && rm -rf  *.tar.gz && chmod a+x /app/http2smtp

# Everything mutable lives under one mountable root: config.ini, the SQLite
# database, rotated application logs, and date-stamped attachment directories.
WORKDIR /data
VOLUME ["/data"]

EXPOSE 12566 12567

# The compatibility page intentionally leaves the public API port closed, so the
# health check verifies the management listener that must be available in every mode.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD nc -z -w 2 127.0.0.1 "${web_port:-12566}" || exit 1

CMD [ "/app/http2smtp","-web_port","12566","-api_port","12567","-storage_root","/data" ]
