# BPF

> There are two types of BPF: cBPF and eBPF (classical and extended). Here we only discuss eBPF and refer to it just as BPF.

BPF is a highly efficient sandboxed virtual machine in the Linux kernel and a BPF program is a program which can be run on this virtual machine. To understand this it is important to understand that the Linux kernel is fundamentally event driven; processes do system calls, hardware sends interrupts etc. and the kernel handles these events. What BPF allows you to do is to trigger a BPF program to run on this virtual machine in response to an event. There are different types of BPF programs which determines, amongst other things, which events will trigger the program to run. BPF therefore makes the Linux kernel programmable at run time.

In addition to being able to run these programs, BPF also has a concept of maps. BPF programs can read and write data to these maps which allows them to communicate with each other. BPF maps are also accessible from user space which could be used, for example, to configure the programs or retrieve useful data such as metrics.

Further information on what BPF provides can be found through the following links:
- [A thorough introduction to eBPF](https://lwn.net/Articles/740157/)
- [How to Make Linux Microservice-Aware with Cilium and eBPF](https://www.infoq.com/presentations/linux-cilium-ebpf)
- [Dive into BPF: a list of reading materials](https://qmonnet.github.io/whirl-offload/2016/09/01/dive-into-bpf/)
- [BPF and XDP Reference Guide](https://docs.cilium.io/en/stable/bpf/)

## Writing BPF programs

BPF defines an [instruction set](https://en.wikipedia.org/wiki/Instruction_set_architecture) which the BPF virtual machine implements. The official documentation for this instruction set can be found in the Linux kernel [networking/filter.txt](https://www.kernel.org/doc/Documentation/networking/filter.txt) documentation and an unofficial reference in IO Visor's [bpf-docs](https://github.com/iovisor/bpf-docs/blob/master/eBPF.md) repository. To write a BPF program we need to write a program using this instruction set. We can then load it into the kernel so that it can be run on the BPF virtual machine.

It is of course possible to write these programs by hand but the easiest way is to make use of the [LLVM](https://en.wikipedia.org/wiki/LLVM) compiler suite which provides a BPF back end. Using [Clang's](https://clang.llvm.org/) LLVM front end then allows us to write BPF programs in C. For a general introduction to this process see [IR is better than assembly](https://idea.popcount.org/2013-07-24-ir-is-better-than-assembly/). A real example of how to use these tools to compile a BPF program and store it in an [ELF](https://en.wikipedia.org/wiki/Executable_and_Linkable_Format) object file can be found in [bpf/Makefile](../bpf/Makefile). ELF is a convenient format for storing BPF programs and maps with each section containing a different program or map.

## Loading BPF programs

Once we have written a BPF program we need to load it into the kernel. This is done using the [`bpf`](http://man7.org/linux/man-pages/man2/bpf.2.html) syscall. This repository makes use of the [gobpf-elf](https://github.com/iovisor/gobpf/tree/master/elf) package which provides the Go bindings for loading and interacting with BPF programs and maps stored in an ELF object file. Examples of how to use this package can be found in [cmd](../cmd).

## TODO

- helpers
- verifier
- pinning
- tail calls
- jit compiler
- comparison with bcc (ldd etc.)