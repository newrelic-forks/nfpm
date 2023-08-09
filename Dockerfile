FROM alpine:3.18.3
COPY nfpm /usr/local/bin/nfpm
ENTRYPOINT ["/usr/local/bin/nfpm"]
