package bpf

import (
	"context"
	"fmt"
	"log"
	"net"
	"unsafe"

	"github.com/iovisor/gobpf/elf"
	cri "github.com/kubernetes/cri-api/pkg/apis/runtime/v1alpha2"
	"google.golang.org/grpc"
)

/*
#include "../../vendor/github.com/iovisor/gobpf/elf/include/bpf.h"
#include "../../bpf/include/exec_trace.h"
*/
import "C"

const (
	execTraceAssetName          = "exec_trace.o"
	execTraceCgroupPath         = "/sys/fs/cgroup/unified"
	execTraceProgramSectionName = "sockops/test"
	bufferSize                  = 256
)

type kSysWrite struct {
	Count     uint32
	Pid       uint32
	Buffer    [bufferSize]byte
	Timestamp uint64
}

type runtimeServer struct{}

func (r *runtimeServer) Version(ctx context.Context, in *cri.VersionRequest) (*cri.VersionResponse, error) {
	fmt.Print("Version\n")
	return &cri.VersionResponse{}, nil
}
func (r *runtimeServer) RunPodSandbox(ctx context.Context, in *cri.RunPodSandboxRequest) (*cri.RunPodSandboxResponse, error) {
	fmt.Print("RunPodSandbox\n")
	return &cri.RunPodSandboxResponse{}, nil
}
func (r *runtimeServer) StopPodSandbox(ctx context.Context, in *cri.StopPodSandboxRequest) (*cri.StopPodSandboxResponse, error) {
	fmt.Print("StopPodSandbox\n")
	return &cri.StopPodSandboxResponse{}, nil
}
func (r *runtimeServer) RemovePodSandbox(ctx context.Context, in *cri.RemovePodSandboxRequest) (*cri.RemovePodSandboxResponse, error) {
	fmt.Print("RemovePodSandbox\n")
	return &cri.RemovePodSandboxResponse{}, nil
}
func (r *runtimeServer) PodSandboxStatus(ctx context.Context, in *cri.PodSandboxStatusRequest) (*cri.PodSandboxStatusResponse, error) {
	fmt.Print("PodSandboxStatus\n")
	return &cri.PodSandboxStatusResponse{}, nil
}
func (r *runtimeServer) ListPodSandbox(ctx context.Context, in *cri.ListPodSandboxRequest) (*cri.ListPodSandboxResponse, error) {
	fmt.Print("ListPodSandbox\n")
	return &cri.ListPodSandboxResponse{}, nil
}
func (r *runtimeServer) CreateContainer(ctx context.Context, in *cri.CreateContainerRequest) (*cri.CreateContainerResponse, error) {
	fmt.Print("CreateContainer\n")
	return &cri.CreateContainerResponse{}, nil
}
func (r *runtimeServer) StartContainer(ctx context.Context, in *cri.StartContainerRequest) (*cri.StartContainerResponse, error) {
	fmt.Print("StartContainer")
	return &cri.StartContainerResponse{}, nil
}
func (r *runtimeServer) StopContainer(ctx context.Context, in *cri.StopContainerRequest) (*cri.StopContainerResponse, error) {
	fmt.Print("StopContainer\n")
	return &cri.StopContainerResponse{}, nil
}
func (r *runtimeServer) RemoveContainer(ctx context.Context, in *cri.RemoveContainerRequest) (*cri.RemoveContainerResponse, error) {
	fmt.Print("RemoveContainer\n")
	return &cri.RemoveContainerResponse{}, nil
}
func (r *runtimeServer) ListContainers(ctx context.Context, in *cri.ListContainersRequest) (*cri.ListContainersResponse, error) {
	fmt.Print("ListContainers\n")
	return &cri.ListContainersResponse{}, nil
}
func (r *runtimeServer) ContainerStatus(ctx context.Context, in *cri.ContainerStatusRequest) (*cri.ContainerStatusResponse, error) {
	fmt.Print("ContainerStatus\n")
	return &cri.ContainerStatusResponse{}, nil
}
func (r *runtimeServer) UpdateContainerResources(ctx context.Context, in *cri.UpdateContainerResourcesRequest) (*cri.UpdateContainerResourcesResponse, error) {
	fmt.Print("UpdateContainerResources\n")
	return &cri.UpdateContainerResourcesResponse{}, nil
}
func (r *runtimeServer) ReopenContainerLog(ctx context.Context, in *cri.ReopenContainerLogRequest) (*cri.ReopenContainerLogResponse, error) {
	fmt.Print("ReopenContainerLog\n")
	return &cri.ReopenContainerLogResponse{}, nil
}
func (r *runtimeServer) ExecSync(ctx context.Context, in *cri.ExecSyncRequest) (*cri.ExecSyncResponse, error) {
	fmt.Print("ExecSync\n")
	return &cri.ExecSyncResponse{}, nil
}
func (r *runtimeServer) Exec(ctx context.Context, in *cri.ExecRequest) (*cri.ExecResponse, error) {
	fmt.Print("Exec: %s\n", in.Cmd)
	return &cri.ExecResponse{}, nil
}
func (r *runtimeServer) Attach(ctx context.Context, in *cri.AttachRequest) (*cri.AttachResponse, error) {
	fmt.Print("Attach\n")
	return &cri.AttachResponse{}, nil
}
func (r *runtimeServer) PortForward(ctx context.Context, in *cri.PortForwardRequest) (*cri.PortForwardResponse, error) {
	fmt.Print("PortForward\n")
	return &cri.PortForwardResponse{}, nil
}
func (r *runtimeServer) ContainerStats(ctx context.Context, in *cri.ContainerStatsRequest) (*cri.ContainerStatsResponse, error) {
	fmt.Print("ContainerStats\n")
	return &cri.ContainerStatsResponse{}, nil
}
func (r *runtimeServer) ListContainerStats(ctx context.Context, in *cri.ListContainerStatsRequest) (*cri.ListContainerStatsResponse, error) {
	fmt.Print("ListContainerStats\n")
	return &cri.ListContainerStatsResponse{}, nil
}
func (r *runtimeServer) UpdateRuntimeConfig(ctx context.Context, in *cri.UpdateRuntimeConfigRequest) (*cri.UpdateRuntimeConfigResponse, error) {
	fmt.Print("UpdateRuntimeConfig\n")
	return &cri.UpdateRuntimeConfigResponse{}, nil
}
func (r *runtimeServer) Status(ctx context.Context, in *cri.StatusRequest) (*cri.StatusResponse, error) {
	fmt.Print("Status\n")
	return &cri.StatusResponse{}, nil
}

func ExecTrace() error {

	// load bpf
	sectionParams := make(map[string]elf.SectionParams)
	sectionParams["maps/ksys_writes"] = elf.SectionParams{PerfRingBufferPageCount: bufferSize}
	m, err := Load(execTraceAssetName, sectionParams)
	if err != nil {
		log.Fatalf("failed to load asset %s: %s", execTraceAssetName, err)
	}

	err = m.EnableKprobes(0)
	if err != nil {
		return fmt.Errorf("failed to enable kprobes: %s", err)
	}
	log.Print("enabled kprobes")

	channel := make(chan []byte)
	lostChannel := make(chan uint64)

	perfMap, err := elf.InitPerfMap(m, "ksys_writes", channel, lostChannel)
	if err != nil {
		return fmt.Errorf("failed to initialise perf map: %s", err)
	}

	perfMap.SetTimestampFunc(kSysWriteTimestamp)
	perfMap.PollStart()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 7777))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	r := runtimeServer{}
	grpcServer := grpc.NewServer()
	cri.RegisterRuntimeServiceServer(grpcServer, &r)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %s", err)
		}
	}()

	var lastPid uint32
	var message []byte

	for {

		select {
		case kSysWrite, ok := <-channel:

			if !ok {
				log.Fatal("channel closed")
			}

			kSysWriteGo := kSysWriteToGo(&kSysWrite)

			currentPid := kSysWriteGo.Pid
			if lastPid == 0 {
				lastPid = currentPid
			}

			if currentPid != lastPid {

				conn, err := net.Dial("tcp", "127.0.0.1:7777")
				if err != nil {
					log.Fatalf("failed to connect to grpc server: %s", err)
				}
				fmt.Print("sending message\n")
				fmt.Fprintf(conn, "%s", message)
				data := make([]byte, 1)
				_, _ = conn.Read(data)
				conn.Close()

				/*response, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					log.Fatalf("failed to retrieve response: %s", err)
				}
				fmt.Printf("response: %s\n", response)*/
				message = []byte{}
				lastPid = currentPid
			} else {
				message = append(message, kSysWriteGo.Buffer[0:kSysWriteGo.Count]...)
			}
			//fmt.Printf("%d\n", kSysWriteGo.Count)
			//fmt.Printf("%s", kSysWriteGo.Buffer[0:kSysWriteGo.Count])

		case _, ok := <-lostChannel:
			if !ok {
				log.Fatal("lost channel closed")
			}
		}
	}

	return nil
}

func kSysWriteTimestamp(data *[]byte) uint64 {
	kSysWrite := (*C.struct_ksys_write_t)(unsafe.Pointer(&(*data)[0]))
	return uint64(kSysWrite.timestamp)
}

func kSysWriteToGo(data *[]byte) (ret kSysWrite) {

	kSysWrite := (*C.struct_ksys_write_t)(unsafe.Pointer(&(*data)[0]))

	ret.Count = uint32(kSysWrite.count)
	ret.Buffer = *(*[C.BUFSIZE]byte)(unsafe.Pointer(&kSysWrite.buf))
	ret.Timestamp = uint64(kSysWrite.timestamp)
	ret.Pid = uint32(kSysWrite.pid)

	return
}
