FROM busybox:1.31.1-glibc

COPY owlet /
COPY buildInfo.json /
COPY i18n /i18n/
ENV GIN_MODE release
ENTRYPOINT ["/owlet"]