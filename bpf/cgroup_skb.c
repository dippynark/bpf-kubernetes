#include <linux/version.h>
#include <uapi/linux/bpf.h>

#include "include/bpf_map.h"
#include "include/bpf_helpers.h"

struct bpf_map_def SEC("maps/count") count_map = {
	.type = BPF_MAP_TYPE_ARRAY,
	.key_size = sizeof(int),
	.value_size = sizeof(__u64),
	.max_entries = 1024,
};

#define DEBUG
#ifdef DEBUG
/* Only use this for debug output. Notice output from bpf_trace_printk()
 * ends up in /sys/kernel/debug/tracing/trace_pipe
 */
#define bpf_debug(fmt, ...)                                                    \
	({                                                                     \
		char ____fmt[] = fmt;                                          \
		bpf_trace_printk(____fmt, sizeof(____fmt), ##__VA_ARGS__);     \
	})
#else
#define bpf_debug(fmt, ...){;}
#endif

SEC("cgroup/skb")
int count_packets(struct __sk_buff *skb) {
    bpf_debug("count_packets\n");

    int packets_key = 0, init_value = 1, ret;
    __u64 *packets = 0;

    packets = bpf_map_lookup_elem(&count_map, &packets_key);
    if (packets == 0)
        return 0;
   
    *packets += 1;

    // allow access
    return 1;
}

char _license[] SEC("license") = "GPL";
__u32 _version SEC("version") = LINUX_VERSION_CODE;