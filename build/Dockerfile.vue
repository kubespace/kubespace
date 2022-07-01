ARG BASEIMAGE=registry.cn-hangzhou.aliyuncs.com/kubespace/node:17.9.0
FROM $BASEIMAGE

COPY ui /app
WORKDIR /app

CMD ["npm", "run", "serve"]
