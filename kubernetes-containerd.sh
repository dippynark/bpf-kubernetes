#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

export DEBIAN_FRONTEND=noninteractive

CONTAINERD_VERSION=1.1.2
CONTAINERD_SOCKET=/var/run/containerd.sock
RUNC_VERSION=1.0.0-rc8

# install kubeadm, kubelet, kubectl and cri-tools
apt-get update && apt-get install -y apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
apt-get update
apt-get install -y kubelet kubeadm kubectl cri-tools
apt-mark hold kubelet kubeadm kubectl cri-tools
echo "runtime-endpoint: unix://$CONTAINERD_SOCKET" > /etc/crictl.yaml

# install containerd
# https://kubernetes.io/docs/setup/cri/#containerd
cat > /etc/sysctl.d/99-kubernetes-cri.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.ipv4.ip_forward                 = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF
sysctl --system
wget -O /usr/local/sbin/runc https://github.com/opencontainers/runc/releases/download/v$RUNC_VERSION/runc.amd64
chmod 755 /usr/local/sbin/runc
wget https://github.com/containerd/containerd/releases/download/v$CONTAINERD_VERSION/containerd-$CONTAINERD_VERSION.linux-amd64.tar.gz
tar xf containerd-$CONTAINERD_VERSION.linux-amd64.tar.gz -C /usr/local
rm containerd-1.1.2.linux-amd64.tar.gz
mkdir -p /etc/containerd
cat <<EOF > /etc/containerd/config.toml
[grpc]
  address = "$CONTAINERD_SOCKET"
  uid = 0
  gid = 0
EOF
curl -o /etc/systemd/system/containerd.service https://raw.githubusercontent.com/containerd/cri/master/contrib/systemd-units/containerd.service
systemctl daemon-reload
systemctl enable containerd
systemctl start containerd

# configure docker
mkdir -p /etc/systemd/system/docker.service.d
cat <<EOF > /etc/systemd/system/docker.service.d/override.conf
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H fd:// --containerd $CONTAINERD_SOCKET
EOF
systemctl daemon-reload
systemctl restart docker

# install kubernetes
kubeadm init --config <(cat <<EOF
apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
nodeRegistration:
  criSocket: $CONTAINERD_SOCKET
EOF
)

# setup kubeconfig
export KUBECONFIG=/etc/kubernetes/admin.conf
chmod 644 "$KUBECONFIG"
cat >/etc/profile.d/kubernetes.sh <<EOF
source /etc/bash_completion
export KUBECONFIG="$KUBECONFIG"
source <(kubectl completion bash)
EOF

# setup kubernetes
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')"
kubectl taint nodes --all node-role.kubernetes.io/master-
kubectl apply -f <(cat <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
EOF
)
