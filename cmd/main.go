package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

	flag.Parse()
	example := *exampleFlag

	switch example {
	case cgroupSkbMetrics:
		err := bpf.CgroupSkbMetrics()
		if err != nil {
			log.Fatalf("failed to run example %s: %s", cgroupSkbMetrics, err)
		}
	default:
		fmt.Printf(helpText, cgroupSkbMetrics)
	}

	os.Exit(0)
}
