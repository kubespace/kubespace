ARG BASEIMAGE=registry.cn-hangzhou.aliyuncs.com/kubespace/distroless-static:latest
FROM $BASEIMAGE

COPY kubespace /
COPY entrypoint.sh /
COPY apps /apps
COPY ui/dist/favicon.ico /favicon.ico

CMD ["/kubespace"]
