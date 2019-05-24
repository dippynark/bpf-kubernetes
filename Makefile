REGISTRY ?= dippynark

DOCKERFILE ?= Dockerfile
IMAGE ?= bpf-kubernetes
TAG ?= $(shell uname -r)

BPF_BUILDER_DOCKERFILE ?= bpf/Dockerfile
BPF_BUILDER_IMAGE ?= bpf-builder

CGROUP2 = /sys/fs/cgroup/unified
DEBUGFS = /sys/kernel/debug

# If you can use docker without being root, you can do "make SUDO="
SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")

run_%:
	@docker run -it \
		--privileged \
		--pid=host \
		--net=host \
		-e NODE_NAME=$(shell hostname) \
		-e KUBECONFIG=$(KUBECONFIG) \
		-v $(KUBECONFIG):$(KUBECONFIG) \
		-v $(CGROUP2):$(CGROUP2) \
		-v $(DEBUGFS):$(DEBUGFS) \
		${REGISTRY}/${IMAGE}:${TAG} \
		--example $*

build: install_bpf docker_build

docker_build:
	$(SUDO) docker build -t ${REGISTRY}/${IMAGE}:${TAG} .

docker_push:
	$(SUDO) docker push ${REGISTRY}/${IMAGE}:${TAG}

install_bpf: bpf
	cp -a $(CURDIR)/bpf/bindata.go $(CURDIR)/pkg/bpf/bindata.go

bpf: docker_build_bpf

docker_build_bpf: docker_build_bpf_image
	$(SUDO) docker run --rm \
		-v $(CURDIR)/bpf:/bpf \
		-v /usr/src:/usr/src \
		--workdir=/bpf \
		$(REGISTRY)/$(BPF_BUILDER_IMAGE) \
		make bindata.go

docker_build_bpf_image:
	$(SUDO) docker build -t $(REGISTRY)/$(BPF_BUILDER_IMAGE) -f $(BPF_BUILDER_DOCKERFILE) ./bpf
