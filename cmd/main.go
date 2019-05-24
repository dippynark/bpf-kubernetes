package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/golang/glog"
	"k8s.io/klog"

	"github.com/dippynark/bpf-kubernetes/pkg/bpf"
)

/*
#include "../vendor/github.com/iovisor/gobpf/elf/include/bpf.h"
*/
import "C"

const (
	cgroupSkbMetrics = "cgroup_skb_metrics"
	helpText         = `Choose from the following examples:
%s
`
)

var exampleFlag = flag.String("example", "", "example to run")

func main() {

	if err := flag.Set("alsologtostderr", "true"); err != nil {
		panic(fmt.Sprintf("failed to set flag alsologtostderr: %s", err))
	}
	if err := flag.Set("v", "2"); err != nil {
		panic(fmt.Sprintf("errors setting flag v: %s", err))
	}
	flag.Parse()

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		if f1.Name == "log_backtrace_at" {
			return
		}
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			if err := f2.Value.Set(value); err != nil {
				panic(fmt.Sprintf("Error applying flag '%s': %v", f1.Name, err))
			}
		}
	})
	defer klog.Flush()

	example := *exampleFlag

	switch example {
	case cgroupSkbMetrics:
		err := bpf.CgroupSkbMetrics()
		if err != nil {
			klog.Fatalf("failed to run example %s: %s", cgroupSkbMetrics, err)
		}
	default:
		klog.Infof(helpText, cgroupSkbMetrics)
	}

	os.Exit(0)
}
