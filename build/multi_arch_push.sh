#!/bin/bash

# 说明：将amd64与arm64镜像合并为一个镜像，并推送到镜像仓库
# 前提：镜像仓库中已存在amd64、arm64镜像

# Usage: ./multi_arch_push.sh <image> <amd64_image> <arm64_image>
#
multi_arch_image=$1
amd64_image=$2
arm64_image=$3

docker manifest create $multi_arch_image $amd64_image $arm64_image --amend
docker manifest annotation $multi_arch_image $arm64_image --arch arm64
docker manifest push $multi_arch_image
