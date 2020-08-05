FROM alpine:3.7
LABEL MAINTAINER="Ted <ski2per@gmail.com>"

RUN sed -i s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g /etc/apk/repositories && \
    apk --no-cache add iptables

COPY dist/flanneld-amd64 /usr/bin/flanneld

ENTRYPOINT ["flanneld"]
