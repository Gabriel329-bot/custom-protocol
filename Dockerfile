FROM alpine:3.19

ARG BINARY_NAME=customproto

RUN apk add --no-cache ca-certificates curl iptables

COPY customproto /usr/local/bin/customproto
RUN chmod +x /usr/local/bin/customproto

EXPOSE 51820/udp 443/tcp

RUN mkdir -p /app

COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["server"]

LABEL maintainer="custom-protocol"
LABEL version="1.0"
LABEL description="Custom Protocol VPN - AmneziaWG + TLS + Custom Handshake"