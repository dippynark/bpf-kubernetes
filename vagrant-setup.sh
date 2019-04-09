#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

export DEBIAN_FRONTEND=noninteractive

apt-get update -y
apt-get install -y \
  linux-image-4.18.0-16-generic \
  linux-headers-4.18.0-16-generic \
  make \
  docker.io \
  llvm

usermod -aG docker vagrant

# disable net_prio and net_cls controllers
# https://elixir.bootlin.com/linux/v4.18/source/include/linux/cgroup-defs.h#L735
sed -ie 's/GRUB_CMDLINE_LINUX=.*/GRUB_CMDLINE_LINUX="cgroup_no_v1=net_prio,net_cls"/' /etc/default/grub
grub-mkconfig -o /boot/grub/grub.cfg
