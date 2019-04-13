# cgroup_skb_metrics

Example cgroup_skb_metrics does the following:
- Loads [cgroup_skb_metrics.c](../bpf/cgroup_skb_metrics.c) containing a `BPF_MAP_TYPE_ARRAY` map (section `maps/count`) and a `BPF_PROG_TYPE_CGROUP_SKB` program (section `cgroup/skb`)
- Creates a v2 cgroup
- Moves itself into the cgroup by writing its PID to the `cgroup.procs` file in the cgroup
- Attaches program `cgroup/skb` to the cgroup using attach type `BPF_CGROUP_INET_EGRESS`
- Sets up an interrupt handler to catch TERM signals and detach program `cgroup/skb` from the cgroup
- Starts generating network traffic in the background
- Regularly checks the count map to retrieve metrics

This has the effect of triggering the BPF program whenever an [IPv4 or IPv6](https://elixir.bootlin.com/linux/v4.18/source/kernel/bpf/cgroup.c#L491) packet egresses from processes in the cgroup. The program has access to all the packet contents but in this case all it does is increment a counter in the count map which is available to user space. Of course, this program could be extended to gather much more interesting data than just packet count, but it demonstrates a nice possibility.

Note that it was important to [disable some v1 controllers](../vagrant-setup.sh) to get this to work. 

## Kubernetes

BPF only supports attaching programs to v2 cgroups. Cgroups v2 is significantly different to cgroups v1; see the `Issues with v1 and Rationales for v2` section of the [official cgroup v2 documentation](https://www.kernel.org/doc/Documentation/cgroup-v2.txt) for a summary as well as [this](https://www.youtube.com/watch?v=P6Xnm0IhiSo) great talk by Christian Brauner at Container Camp 2017. Most container runtimes that I know of that make use of cgroups for resource accounting and management only support v1. The main reason for this is that cgroups v2 is not at feature parity with cgroups v1; many key v1 controllers have not been implemented yet for cgroups v2 (see [this](https://github.com/opencontainers/runc/issues/654) discussion from a `runc` issue). Once these key controllers are implemented and support for cgroups v2 is added to more container runtimes that integrate with Kubernetes, BPF programs can be attached to the cgroups set up by Kubernetes and provide a huge number of new possibilities for monitoring/security etc.

Right now, if you had a way of mirroring the Kubernetes cgroup v1 hierarchy from and below the [kubepods](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/node/node-allocatable.md#recommended-cgroups-setup) cgroup in the cgroup v2 hierarchy you could write a Kubernetes controller to attach BPF programs such as the one above to each cgroup for each Pod and container and expose the collected data through a Prometheus endpoint.

## Debugging

This section contains issues I came across and how I debugged them; maybe someone will find them useful.

### Program not working inside a container

By tailing debug output `sudo cat /sys/kernel/debug/tracing/trace_pipe` I noticed my program only worked outside of a Docker container. You can find the problem in more detail on [Stack Overflow](https://stackoverflow.com/questions/55646983/why-does-my-bpf-prog-type-cgroup-skb-program-not-work-in-a-container) as well as the solution (to disable the `net_prio` and `net_cls` v1 controllers).

I got to this by finding out where the BPF program was being called from ([here](https://elixir.bootlin.com/linux/v4.18/source/net/ipv4/ip_output.c#L297)), following seemingly related calls through ([here](https://elixir.bootlin.com/linux/v4.18/source/include/linux/bpf-cgroup.h#L85), [here](https://elixir.bootlin.com/linux/v4.18/source/kernel/bpf/cgroup.c#L515) and [here](https://elixir.bootlin.com/linux/v4.18/source/include/linux/cgroup.h#L772)) and seeing the comment above the `sock_cgroup_data` struct ([here](https://elixir.bootlin.com/linux/v4.18/source/include/linux/cgroup-defs.h#L735), maybe a bit lucky).

They can be disabled by setting the kernel boot parameter `cgroup_no_v1=net_prio,net_cls`.