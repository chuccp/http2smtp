# syntax=docker/dockerfile:1
FROM busybox:latest
WORKDIR /app

ARG goarch=amd64

ADD https://github.com/chuccp/http2smtp/releases/latest/download/http2smtp-linux-${goarch}.tar.gz /app/
RUN tar -xzf http2smtp-linux-${goarch}.tar.gz && rm -rf  *.tar.gz && chmod a+x /app/http2smtp
EXPOSE 12566 12567
CMD [ "/app/http2smtp","-web_port","12566","-api_port","12577"  ]