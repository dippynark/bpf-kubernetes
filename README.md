# bpf-kubernetes

bpf-kubernetes contains a collection of POC examples of BPF programs that could be used to provide benefits to a Kubernetes cluster operator. A lot of inspiration came from the blog post: [Using eBPF in Kubernetes](https://kubernetes.io/blog/2017/12/using-ebpf-in-kubernetes/).

## Documentation

It's recommended to read the documentation in the same order as the following list; each document assumes all knowledge from any previous documents:

- General BPF documentation can be found in [bpf.md](docs/bpf.md)
- Documentation specific to each example can be found below:
  - [cgroup_skb_metrics.md](docs/cgroup_skb_metrics.md)
  - [exec_trace.md](docs/exec_trace.md)

## Quickstart

[Vagrant](https://www.vagrantup.com/) can be used to spin up a virtual environment to build and load the BPF programs. The environment depends on [VirtualBox](https://www.virtualbox.org/wiki/Downloads) but other [providers](https://www.vagrantup.com/docs/providers/) exist.

```
$ vagrant plugin install vagrant-reload vagrant-disksize
$ vagrant box list | grep ubuntu/bionic64 || vagrant box add ubuntu/bionic64
$ vagrant up
$ vagrant ssh
$ cd /vagrant
$ make build
$ make run_help
Choose from the following examples:
cgroup_skb_metrics
$ make run_cgroup_skb_metrics
I0524 16:12:18.585097   29216 cgroup_skb_metrics.go:247] packet count: 0
I0524 16:12:18.586693   29216 cgroup_skb_metrics.go:155] starting controller
I0524 16:12:18.688733   29216 cgroup_skb_metrics.go:124] attached bpf program to pod curl-b7ff57fc5-tjm42 container curl cgroup
I0524 16:12:23.590334   29216 cgroup_skb_metrics.go:247] packet count: 2
I0524 16:12:28.591552   29216 cgroup_skb_metrics.go:247] packet count: 15
I0524 16:12:33.592311   29216 cgroup_skb_metrics.go:247] packet count: 28
^C
```

## Ideas

- Prevent a kubectl exec session from connecting out over the network
- Capture all kubectl exec output for auditing purpose
- https://cilium.io/blog/2018/10/23/cilium-13-envoy-go/
