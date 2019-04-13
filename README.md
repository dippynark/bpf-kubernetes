# bpf-kubernetes

bpf-kubernetes contains a collection of POC examples of BPF programs that could be used to provide benefits to a Kubernetes cluster. A lot of inspiration came from the blog post: [Using eBPF in Kubernetes](https://kubernetes.io/blog/2017/12/using-ebpf-in-kubernetes/).

## Documentation

It's recommended to read the documentation in the same order as the following list; each document assumes all knowledge from any previous documents:

- General BPF documentation can be found in [bpf.md](docs/bpf.md)
- Documentation specific to each example can be found below:
  - [cgroup_skb_metrics.md](docs/cgroup_skb_metrics.md)

## Quickstart

[Vagrant](https://www.vagrantup.com/) can be used to spin up a virtual environment to build and load the BPF programs. The environment depends on [VirtualBox](https://www.virtualbox.org/wiki/Downloads) but other [providers](https://www.vagrantup.com/docs/providers/) exist.

```
$ vagrant plugin install vagrant-reload
$ vagrant box list | grep ubuntu/bionic64 || vagrant box add ubuntu/bionic64
$ vagrant up
$ vagrant ssh
$ cd /vagrant
$ make build
$ make run_help
Choose from the following examples:
cgroup_skb_metrics
$ make run_cgroup_skb_metrics
2019/04/13 16:40:12 packet count: 0
2019/04/13 16:40:16 pinged www.google.com
2019/04/13 16:40:17 packet count: 3
2019/04/13 16:40:18 curled http://www.google.com/
2019/04/13 16:40:20 pinged www.google.com
2019/04/13 16:40:22 packet count: 13
2019/04/13 16:40:22 curled http://www.google.com/
2019/04/13 16:40:26 pinged www.google.com
2019/04/13 16:40:27 packet count: 23
^C
```

## Ideas

- Prevent a kubectl exec session from connecting out over the network
- Capture all kubectl exec output for auditing purpose
- https://cilium.io/blog/2018/10/23/cilium-13-envoy-go/
