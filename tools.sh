#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

export DEBIAN_FRONTEND=noninteractive

apt-get update && apt-get install -y \
  make \
  llvm \
  docker.io

# add vagrant user to docker group
usermod -aG docker vagrant

# use systemd cgroup driver
# https://kubernetes.io/docs/setup/cri/#docker
cat > /etc/docker/daemon.json <<EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
systemctl daemon-reload
systemctl restart docker
