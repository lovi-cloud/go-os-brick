FROM whywaita/gotgt

MAINTAINER whywaita <https://github.com/whywaita>

RUN apt install -y open-iscsi multipath-tools

EXPOSE 3260

CMD ["/go/src/github.com/gostor/gotgt/_output/cmd/bin/gotgt", "daemon", "--log", "debug"]