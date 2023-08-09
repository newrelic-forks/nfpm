FROM alpine:3.18.3
ARG package
COPY ${package} /tmp/foo.apk
RUN apk add --allow-untrusted /tmp/foo.apk
# TODO: seems like there is no changelog support on apk
