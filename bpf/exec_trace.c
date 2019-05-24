// error: unknown type name 'atomic64_t'; did you mean 'atomic_t'?
#include <linux/kconfig.h>

#include <linux/skbuff.h>
#include <linux/netdevice.h>
#include <uapi/linux/bpf.h>
#include <linux/version.h>
#include <linux/un.h>
#include <linux/fs.h>

#include "include/bpf_helpers.h"
#include "include/bpf_map.h"
#include "include/exec_trace.h"

struct bpf_map_def SEC("maps/pid_fds") pid_fds = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(struct pid_fd_t),
	.value_size = sizeof(int),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

// define maps
struct bpf_map_def SEC("maps/ksys_writes") ksys_writes = {
	.type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
	.key_size = sizeof(int),
	.value_size = sizeof(__u32),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

#define DEBUG
#ifdef DEBUG
/* Only use this for debug output. Notice output from bpf_trace_printk()
 * ends up in /sys/kernel/debug/tracing/trace_pipe
 */
#define bpf_debug(fmt, ...)                                        \
	({                                                               \
		char ____fmt[] = fmt;                                          \
		bpf_trace_printk(____fmt, sizeof(____fmt), ##__VA_ARGS__);     \
	})
#else
#define bpf_debug(fmt, ...){;}
#endif

// https://elixir.bootlin.com/linux/latest/source/net/socket.c#L1674
SEC("kprobe/__sys_connect")
int bpf_prog1(struct pt_regs *ctx) {

	struct sockaddr_un sockaddr_un = {};
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubelet/app/options/options.go#L217
	char socket[] = "/var/run/containerd.sock";
	char *sun_path;
	int fd = (int)PT_REGS_PARM1(ctx);
	void *sockaddr_arg = (void *)PT_REGS_PARM2(ctx);
	int ret, sockaddr_len_arg = (int)PT_REGS_PARM3(ctx);
	struct pid_fd_t pid_fd = {};

	if (sockaddr_len_arg > sizeof(sockaddr_un)) {
		//bpf_debug("argument sockaddr too large for sockaddr_un: %d\n");
		return 0;
	}

	ret = bpf_probe_read(&sockaddr_un, sizeof(sockaddr_un), sockaddr_arg);
	if (ret != 0) {
		bpf_debug("ERROR: failed to read sockaddr_un from sockaddr_arg: %d\n", ret);
		return 0;
	}

	if (sockaddr_un.sun_family != AF_UNIX) {
		return 0;
	}
	sun_path = sockaddr_un.sun_path;

	for (int i = 0; i < sizeof(socket); i++) {
		if (socket[i] != sun_path[i]) {
			return 0;
		}
	}

	pid_fd.pid = bpf_get_current_pid_tgid();
	pid_fd.fd = fd;
	bpf_debug("socket connection found for process %d: %d\n", pid_fd.pid, pid_fd.fd);

	ret = bpf_map_update_elem(&pid_fds, &pid_fd, &fd, BPF_ANY);
	if (ret < 0) {
		bpf_debug("ERROR: failed to update element: %d\n", ret);
		return 0;
	}

	return 0;
}

// https://elixir.bootlin.com/linux/v4.18.20/source/fs/read_write.c#L610
SEC("kprobe/ksys_write")
int bpf_prog2(struct pt_regs *ctx) {

	struct ksys_write_t ksys_write = {};
	int *fd_pointer, pid, i, ret;
	int fd = (int)PT_REGS_PARM1(ctx);
	void *buf_arg = (void *)PT_REGS_PARM2(ctx);
	size_t count_arg = (size_t)PT_REGS_PARM3(ctx);
	struct pid_fd_t pid_fd = {};

	pid_fd.pid = bpf_get_current_pid_tgid();
	pid_fd.fd = fd;
	ksys_write.pid = pid_fd.pid;

	fd_pointer = (int *)bpf_map_lookup_elem(&pid_fds, &pid_fd);
	if (!fd_pointer) {
		return 0;
	}

	bpf_debug("ksys_write found: %d %d\n", pid_fd.pid, pid_fd.fd);

	/*if (fd_arg != *fd_pointer) {
		return 0;
	}*/

	#pragma unroll
	for (i = 0; i < LOOPSIZE && i < count_arg / BUFSIZE; i++) {
		ret = bpf_probe_read(&ksys_write.buf, BUFSIZE, buf_arg + (i * BUFSIZE));
		if (ret != 0) {
			bpf_debug("ERROR: failed to read BUFSIZE from buf_arg: %d\n", ret);
			return 0;
		}
		ksys_write.count = BUFSIZE;
		ksys_write.timestamp = bpf_ktime_get_ns();
		ret = bpf_perf_event_output(ctx, &ksys_writes, 0, &ksys_write, sizeof(ksys_write));
		if (ret != 0) {
			bpf_debug("ERROR: failed to read count_arg %% BUFSIZE from buf_arg: %d\n", ret);
			return 0;
		}
	}

	if (i == LOOPSIZE) {
		bpf_debug("ERROR: write too big: %d\n", count_arg);
	}

	ret = bpf_probe_read(&ksys_write.buf, count_arg % BUFSIZE, buf_arg + (i * BUFSIZE));
	if (ret != 0) {
		bpf_debug("ERROR: failed to read count_arg %% BUFSIZE from buf_arg: %d\n", ret);
		return 0;
	}
	ksys_write.count = count_arg % BUFSIZE;
	ksys_write.timestamp = bpf_ktime_get_ns();
	ret = bpf_perf_event_output(ctx, &ksys_writes, 0, &ksys_write, sizeof(ksys_write));
	if (ret != 0) {
		bpf_debug("ERROR: failed to read count_arg %% BUFSIZE from buf_arg: %d\n", ret);
		return 0;
	}

	return 0;
}

/*SEC("kprobe/ksys_read")
int bpf_prog3(struct pt_regs *ctx) {

	char buf[BUFSIZE + 1];
	int *fd_pointer, pid, fd_arg = PT_REGS_PARM1(ctx);
	void *buf_arg = (void *)PT_REGS_PARM2(ctx);
	size_t count_arg = (size_t)PT_REGS_PARM3(ctx);

	pid = bpf_get_current_pid_tgid();

	fd_pointer = (int *)bpf_map_lookup_elem(&pid_fd, &pid);
	if (!fd_pointer) {
		return 0;
	}

	if (fd_arg != *fd_pointer) {
		return 0;
	}

	if (count_arg > 4096) {
		count_arg = 4096;
	}

	// zero out buffer
	for (int i = 0; i < BUFSIZE + 1; i++) {
		buf[i] = 0;
	}

	int limit = 15;
	for (int i = 0; i < limit; i++) {
		bpf_probe_read(buf, BUFSIZE, buf_arg + (i * BUFSIZE));
		bpf_debug("%s\n", buf);
	}

	for (int i = 0; i < BUFSIZE + 1; i++) {
		buf[i] = 0;
	}

	bpf_probe_read(buf, count_arg % BUFSIZE, buf_arg);
	bpf_debug("%s\n", buf);

	return 0;
}*/

char _license[] SEC("license") = "GPL";
__u32 _version SEC("version") = LINUX_VERSION_CODE;