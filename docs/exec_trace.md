# exec_trace

> This is currently just a proposal

Example exec_trace will capture all output from Docker exec sessions. This should also work on Kubernetes clusters using Docker.

## Current approach

When considering how to access kubectl exec sessions, we need to attach to a point in the kernel that will have access to the unencrypted session. Watching kubelet's HTTPS endpoint is therefore not an option; all content is encrypted/decrypted in user space.

However, the docker socket (/var/run/docker.sock) is unencrypted by default (https://tech.paulcz.net/blog/secure-docker-with-tls/) and either way that is under our control as we can rely on the security of having to have file access to the socket.

It's nice that this should work for docker CLI exec sessions too in the same way.

So use a BPF_PROG_TYPE_SOCK_OPS program to watch for UNIX domain socket connections to the docker socket

Set a socket mark on the socket to catch relevant packets

Use a BPF_PROG_TYPE_CGROUP_SKB program with attach type Ingress to look for packets coming into that socket (filtering on the mark field on the context `struct __sk_buff *`)

Parse unencrypted protocol (might be hard)