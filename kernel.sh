#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

export DEBIAN_FRONTEND=noninteractive

apt-get update && apt-get install -y \
  linux-image-4.18.0-17-generic \
  linux-headers-4.18.0-17-generic

# enable hybrid cgroup mode and disable net_prio and net_cls controllers
# https://github.com/kinvolk/inspektor-gadget/blob/master/Documentation/install.md#on-another-kubernetes-distribution
# https://elixir.bootlin.com/linux/v4.18/source/include/linux/cgroup-defs.h#L735
sed -ie 's/GRUB_CMDLINE_LINUX=.*/GRUB_CMDLINE_LINUX="systemd.unified_cgroup_hierarchy=false systemd.legacy_systemd_cgroup_controller=false cgroup_no_v1=net_prio,net_cls"/' /etc/default/grub
grub-mkconfig -o /boot/grub/grub.cfg
