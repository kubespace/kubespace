#!/bin/sh

set -eu
set -o pipefail

readonly SYSTEMD_UNIT_PATH="/usr/lib/systemd/system"
readonly UNIT_FILE="spacelet.service"
readonly BINARY_PATH="/usr/local/bin"
readonly HTTP_SERVER=http://{{ .ServerHost }}
readonly SPACELET_PORT={{ .Port }}
readonly SPACELET_DATADIR={{ .DataDir }}
readonly SPACELET_HOSTIP={{ .HostIp }}

HostIpVar=
if [ ${SPACELET_HOSTIP} != '' ]; then
  HostIpVar="--host-ip ${SPACELET_HOSTIP}"
fi

EXIT_CODE=0

# 检查端口是否被占用
(lsof -i:${SPACELET_PORT} -sTCP:LISTEN +c0 || true) | awk '{if ((NR>1 && $1!="spacelet")) exit 1}' || EXIT_CODE=1
if [ $EXIT_CODE -eq 1 ]; then
  echo -e "\033[31m spacelet安装失败，${SPACELET_PORT}端口已经被占用 \033[0m"
  exit 1
fi

# 安装spacelet
cd /tmp
rm -rf spacelet && mkdir spacelet
cd spacelet
wget ${HTTP_SERVER}/api/v1/assets/{{ .OS }}/{{ .Arch }}/spacelet -O spacelet

mkdir -p ${BINARY_PATH}
install -m 755 spacelet ${BINARY_PATH}/

cat << EOF > ${UNIT_FILE}
[Unit]
Description=Spacelet Service
ConditionPathExists=/usr/local/bin/spacelet
After=network.target

[Service]
Type=simple
User=root
Group=root
# Limit the number of open files to avoid exhaustion
LimitNOFILE=4096

ExecStart=/usr/local/bin/spacelet --server-url ${HTTP_SERVER} --port ${SPACELET_PORT} --data-dir ${SPACELET_DATADIR} ${HostIpVar}
# When kill the service timeout, it will be kill -9
# TimeoutSec = TimeoutStartSec and TimeoutStopSec
TimeoutSec=10
# Always restart the service
Restart=always
RestartSec=10

[Install]
EOF

mkdir -p ${SYSTEMD_UNIT_PATH}
install -m 644 -b ${UNIT_FILE} ${SYSTEMD_UNIT_PATH}/

systemctl enable ${UNIT_FILE}
systemctl restart ${UNIT_FILE}

cd /tmp
rm -rf spacelet

# 检查spacelet是否启动
sleep 1

echo -e '\033[32m spacelet安装并启动成功 \033[0m'
