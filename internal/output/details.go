package output

import (
	"context"
	"fmt"
	"github.com/cen-ngc5139/nfs-trace/internal"
	ebpfbinary "github.com/cen-ngc5139/nfs-trace/internal/binary"
	"github.com/cen-ngc5139/nfs-trace/internal/metadata"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/perf"
	"k8s.io/klog/v2"
	"os"
	"time"
)

func ProcessEvents(coll *ebpf.Collection, ctx context.Context, addr2name internal.Addr2Name) {
	events := coll.Maps["rpc_task_map"]
	// Set up a perf reader to read events from the eBPF program
	rd, err := perf.NewReader(events, os.Getpagesize())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Creating perf reader failed: %v\n", err)
		os.Exit(1)
	}
	defer rd.Close()

	fmt.Printf("Addr \t\t PID \t\t Pod Name \t\t Container ID \t\t Mount \t\t NFS Mount \t\t File \t\t MountID \n")
	var event ebpfbinary.KProbePWRURpcTaskFields
	for {
		for {
			if err := parseEvent(rd, &event); err == nil {
				break
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Microsecond):
				continue
			}
		}

		mountList, err := metadata.ParseMountInfo(fmt.Sprintf("/proc/%d/mountinfo", event.Pid))
		if err != nil {
			klog.Errorf("Failed to get mount info: %v", err)
			continue
		}

		mountInfo, err := metadata.GetMountInfoFormObj(fmt.Sprintf("%d", event.MountId), mountList)
		if err != nil {
			klog.Errorf("Failed to get mount info: %v", err)
			continue
		}

		funcName := addr2name.FindNearestSym(event.CallerAddr)
		fmt.Printf("%s \t\t%d \t\t%s \t\t%s \t\t%s \t\t%s \t\t%s \t\t%d \n",
			funcName, event.Pid, convertInt8ToString(event.Pod[:]), convertInt8ToString(event.Container[:]),
			mountInfo.LocalMountDir, mountInfo.RemoteNFSAddr, parseFileName(event.File[:]), event.MountId)

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
