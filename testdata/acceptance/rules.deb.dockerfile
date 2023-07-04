FROM ubuntu:22.10
ARG package
COPY ${package} /tmp/foo.deb
RUN dpkg -i /tmp/foo.deb
RUN dpkg -r foo
