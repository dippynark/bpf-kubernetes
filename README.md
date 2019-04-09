# bpf-kubernetes

bpf-kubernetes contains a collection of POC examples of BPF programs that could be used to provide benefits to a Kubernetes cluster. A lot of inspiration came from the blog post: [Using eBPF in Kubernetes](https://kubernetes.io/blog/2017/12/using-ebpf-in-kubernetes/).

## Documentation

- General BPF documentation can be found in [bpf.md](docs/bpf.md)
- Documentation specific to each example can be found below:
  - [cgroup_skb.md](docs/cgroup_skb.md)

## Quickstart

[Vagrant](https://www.vagrantup.com/) can be used to spin up a virtual environment to build and load the BPF programs. The environment depends on [VirtualBox](https://www.virtualbox.org/wiki/Downloads) but other [providers](https://www.vagrantup.com/docs/providers/) exist.

```
$ vagrant plugin install vagrant-reload
$ vagrant box list | grep ubuntu/bionic64 || vagrant box add ubuntu/bionic64
$ vagrant up
$ vagrant ssh
$ cd /vagrant
$ make run
```

## Ideas

- BPF program to prevent a kubectl exec from reaching out over the network
- https://cilium.io/blog/2018/10/23/cilium-13-envoy-go/
