ARG BASEIMAGE=registry.cn-hangzhou.aliyuncs.com/kubespace/node:17.9.0
FROM $BASEIMAGE

COPY . /app
WORKDIR /app

CMD ["npm", "run", "serve"]
