ARG BASEIMAGE=openspacee/ospserver-base:latest
FROM $BASEIMAGE

COPY kubespace /
COPY entrypoint.sh /
COPY apps /apps
COPY ui/dist/favicon.ico /favicon.ico

CMD ["bash", "-c", "sh /entrypoint.sh"]
