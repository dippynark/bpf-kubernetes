# Headers:
# - ./include/bpf_helpers.h: tools/testing/selftests/bpf/bpf_helpers.h
# - ./include/bpf_endian.h: tools/testing/selftests/bpf/bpf_endian.h
# - ./include/bpf_map.h; https://github.com/iovisor/gobpf/blob/master/elf/include/bpf_map.h
# - ./include/linux/types.h: tools/testing/selftests/bpf/include/uapi/linux/types.h
#
# Assumptions:
# - linux headers in /usr/src/linux-headers-`uname -r`
# - `struct bpf_map_def` in bpf_helpers.h has been commented out to use definition in bpf_map.h
# - llc/clang are installed and support "bpf" target; check with "llc --version"
# - go-bindata is installed: https://github.com/jteeuwen/go-bindata

SHELL=/bin/bash -o pipefail

OBJS ?= cgroup_skb_metrics.o

LINUX_HEADERS ?= /usr/src/linux-headers-$(shell uname -r)
LINUX_INCLUDE = -I$(LINUX_HEADERS)/arch/x86/include \
								-I$(LINUX_HEADERS)/arch/x86/include/generated \
								-I$(LINUX_HEADERS)/include \
								-I$(LINUX_HEADERS)/include/generated/uapi \
								-I$(LINUX_HEADERS)/arch/x86/include/uapi \
								-I$(LINUX_HEADERS)/include/uapi

CLANG ?= clang
# -nostdinc: fatal error: 'asm/unistd_64.h' file not found
CLANG_FLAGS ?= -isystem $(shell $(CLANG) -print-file-name=include) \
	-D__BPF_TRACING__ -D__KERNEL__ -D__ASM_SYSREG_H \
	-D__NR_CPUS__=$(shell nproc) \
	-Wno-unused-value -Wno-pointer-sign \
	-Wno-compare-distinct-pointer-types \
	-Wno-gnu-variable-sized-type-not-at-end \
	-Wno-address-of-packed-member -Wno-tautological-compare \
	-Wno-unknown-warning-option \
	$(LINUX_INCLUDE) -O2 -emit-llvm

LLC ?= llc
LLC_FLAGS ?= -march=bpf -filetype=obj -relocation-model=ropi

GO_BINDATA ?= go-bindata
GO_BINDATA_FLAGS ?= -pkg bpf -modtime 1

all: $(OBJS)

bindata.go: $(OBJS)
	$(GO_BINDATA) $(GO_BINDATA_FLAGS) -o bindata.go $(OBJS)

.DELETE_ON_ERROR:
%.o: %.c include/%.h
	$(CLANG) $(CLANG_FLAGS) -c $< -o - | $(LLC) $(LLC_FLAGS) -o "$@"

clean:
	rm -f $(OBJS) bindata.go
