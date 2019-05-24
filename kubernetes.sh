#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

export DEBIAN_FRONTEND=noninteractive

# install kubeadm, kubelet and kubectl
# https://kubernetes.io/docs/setup/independent/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl
apt-get update && apt-get install -y apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
apt-get update
apt-get install -y kubelet kubeadm kubectl
apt-mark hold kubelet kubeadm kubectl

# configure cgroup driver
cat > /etc/default/kubelet <<EOF
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd
EOF

kubeadm init

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
  name: curl
  labels:
    app: curl
spec:
  selector:
    matchLabels:
      app: curl
  template:
    metadata:
      labels:
        app: curl
    spec:
      containers:
      - name: curl
        image: pstauffer/curl
        ports:
        - containerPort: 80
        command:
        - sh
        - -c
        - |
          while true; do
            curl http://google.com
            sleep 5
          done
EOF
)
