ARG BASEIMAGE=openspacee/ospserver-base:latest
FROM $BASEIMAGE

COPY ospserver /
COPY entrypoint.sh /
COPY helm_apps /helm_apps
COPY ui/dist/favicon.ico /favicon.ico

CMD ["bash", "-c", "sh /entrypoint.sh"]
