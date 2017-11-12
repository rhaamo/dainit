# vim:set ft=dockerfile:
FROM ubuntu:17.10
#FROM debian:stable

LABEL maintainer="Dashie <dashie@sigpipe.me>"

LABEL org.label-schema.license=MIT \
    org.label-schema.name=lutrainit \
    org.label-schema.vcs-url=https://dev.sigpipe.me/dashie/lutrainit

COPY lutractl/lutractl /usr/local/bin/
COPY lutrainit/lutrainit /usr/local/bin/
COPY conftest /etc/lutrainit

ENTRYPOINT ["/usr/local/bin/lutrainit"]

# No CMD, we are testing an init
#CMD ["asterisk", "-vvf", "-T", "-W", "-U", "asterisk", "-p"]
