#!/bin/sh

docker build -t registry.cn-hangzhou.aliyuncs.com/zgjhub/kubespace-controller:1.1.2 -f Dockerfile_controller .
docker build -t registry.cn-hangzhou.aliyuncs.com/zgjhub/kube-agent:1.1.2 -f Dockerfile_kubeagent .
docker build -t registry.cn-hangzhou.aliyuncs.com/zgjhub/kubespace-server:1.1.2 -f Dockerfile_server .

docker push registry.cn-hangzhou.aliyuncs.com/zgjhub/kubespace-controller:1.1.2
docker push registry.cn-hangzhou.aliyuncs.com/zgjhub/kube-agent:1.1.2
docker push registry.cn-hangzhou.aliyuncs.com/zgjhub/kubespace-server:1.1.2