package bpf

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/iovisor/gobpf/elf"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/sparrc/go-ping"
)

const (
	cgroupSkbMetricsAssetName   = "cgroup_skb_metrics.o"
	cgroupSkbMetricsCgroupPath  = "/sys/fs/cgroup/unified/bpf-kubernetes"
	cgroupSkbProgramSectionName = "cgroup/skb"
	countMapSectionName         = "maps/count"
	mapSectionPrefix            = "maps/"
	defaultFilePerm             = 0700
	pingHost                    = "www.google.com"
	curlAddress                 = "http://www.google.com/"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func CgroupSkbMetrics() error {

	// load bpf
	m, err := Load(cgroupSkbMetricsAssetName)
	if err != nil {
		log.Fatal("failed to load asset %s: %s", cgroupSkbMetricsAssetName, err)
	}

	// create cgroup
	fileInfo, err := os.Stat(cgroupSkbMetricsCgroupPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(cgroupSkbMetricsCgroupPath, defaultFilePerm)
		if err != nil {
			log.Fatalf("failed to create cgroup %s: %s", cgroupSkbMetricsCgroupPath, err)
		}
	} else if !fileInfo.IsDir() {
		log.Fatalf("failed to create cgroup %s: path exists but it is not a directory", cgroupSkbMetricsCgroupPath)
	}

	// add process to cgroup
	err = cgroups.WriteCgroupProc(cgroupSkbMetricsCgroupPath, os.Getpid())
	if err != nil {
		log.Fatalf("failed to write to cgroup %s: %s", cgroupSkbMetricsCgroupPath, err)
	}

	// attach program to cgroup
	cgroupProgram := m.CgroupProgram(cgroupSkbProgramSectionName)
	if cgroupProgram == nil {
		log.Fatalf("failed to retrieve cgroup program %s", cgroupSkbProgramSectionName)
	}
	err = elf.AttachCgroupProgram(cgroupProgram, cgroupSkbMetricsCgroupPath, elf.EgressType)
	if err != nil {
		log.Fatalf("failed to attach cgroup program %s to %s: %s", cgroupSkbProgramSectionName, cgroupSkbMetricsCgroupPath, err)
	}

	// setup interrupt handler
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interruptChan
		err := elf.DetachCgroupProgram(cgroupProgram, cgroupSkbMetricsCgroupPath, elf.EgressType)
		if err != nil {
			log.Fatalf("failed to detach cgroup program %s from %s: %s", cgroupSkbProgramSectionName, cgroupSkbMetricsCgroupPath, err)
		}
		os.Exit(0)
	}()

	// generate ping traffic
	go func() {
		for {
			r := rand.Intn(10)
			time.Sleep(time.Duration(r) * time.Second)
			err = doPing(pingHost)
			if err != nil {
				log.Fatalf("failed to ping %s: %s", pingHost, err)
			}
			log.Printf("pinged %s", pingHost)
			time.Sleep(time.Second)
		}
	}()

	// generate curl traffic
	go func() {
		for {
			r := rand.Intn(10)
			time.Sleep(time.Duration(r) * time.Second)
			err = doCurl(curlAddress)
			if err != nil {
				log.Fatalf("failed to curl %s: %s", curlAddress, err)
			}
			log.Printf("curled %s", curlAddress)
		}
	}()

	// watch count map
	countMapName := strings.TrimPrefix(countMapSectionName, mapSectionPrefix)
	countMap := m.Map(countMapName)
	for {
		key := 0
		var value uint64
		err = m.LookupElement(countMap, unsafe.Pointer(&key), unsafe.Pointer(&value))
		if err != nil {
			log.Fatalf("failed to lookup map %s: %s", countMapName, err)
		}
		log.Printf("packet count: %d", value)
		time.Sleep(time.Second * 5)
	}

}

func doPing(host string) error {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return err
	}
	pinger.Count = 1
	pinger.SetPrivileged(true)
	pinger.Run()
	return nil
}

func doCurl(address string) error {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request: %s", err)
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %s", err)
	}

	return nil
}
