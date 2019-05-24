#include <linux/types.h>

#define BUFSIZE 256
#define LOOPSIZE 128

struct pid_fd_t {
	__u32 pid;
	int fd;
};

struct ksys_write_t {
    char buf[BUFSIZE];
    __u32 pid;
    __u32 count;
    __u64 timestamp;
};
