# cgroup_skb

[cgroup_skb](../examples/main.go) collects metrics on network traffic egressing from a process in a v2 cgroup.

## Debug

When I initially ran my program, no matching was happening, but if I copied my binary outside of the container it worked.

https://elixir.bootlin.com/linux/v4.18/source/net/ipv4/ip_output.c#L297
https://elixir.bootlin.com/linux/v4.18/source/include/linux/bpf-cgroup.h#L85
https://elixir.bootlin.com/linux/v4.18/source/kernel/bpf/cgroup.c#L515
https://elixir.bootlin.com/linux/v4.18/source/include/linux/cgroup.h#L772
https://elixir.bootlin.com/linux/v4.18/source/include/linux/cgroup-defs.h#L735

https://elixir.bootlin.com/linux/latest/source/kernel/cgroup/cgroup-v1.c#L1290
cgroup_no_v1=net_prio,net_cls