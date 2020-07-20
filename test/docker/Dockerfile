FROM debian:buster

MAINTAINER whywaita <https://github.com/whywaita>

RUN apt update -y -qq \
    && apt install -y targetcli-fb dbus kmod

## targetcli need privileged, so execute when docker run --privileged.
COPY init.sh /

EXPOSE 3260/tcp

ENTRYPOINT ["/init.sh"]