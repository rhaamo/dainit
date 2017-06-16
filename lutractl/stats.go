package main

import (
	"github.com/urfave/cli"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"fmt"
)

var CmdStats = cli.Command {
	Name: "stats",
	Usage: "Shows init stats",
	Description: "Shows init stats",
	Action: getStats,
	Flags: []cli.Flag{},
}

func getStats(ctx *cli.Context) error {
	res, err := GorpcDispatcherClient.Call("stats", nil)

	resIpc := res.(*ipc.IpcSysStatus)

	fmt.Printf("lutrainit running statistics\n")
	fmt.Printf("Init Uptime: %s\n", resIpc.Uptime)
	fmt.Printf("Current Goroutines: %d\n\n", resIpc.NumGoroutine)

	fmt.Printf("Current Memory Usage: %s\n", resIpc.MemAllocated)
	fmt.Printf("Total Memory Allocated: %s\n", resIpc.MemTotal)
	fmt.Printf("Memory Obtained: %s\n", resIpc.MemSys)
	fmt.Printf("Pointer Lookup Times: %d\n", resIpc.Lookups)
	fmt.Printf("Memory Allocate Times: %d\n", resIpc.MemMallocs)
	fmt.Printf("Memory Free Times: %d\n\n", resIpc.MemFrees)

	fmt.Printf("Current Heap Usage: %s\n", resIpc.HeapAlloc)
	fmt.Printf("Heap Memory Obtained: %s\n", resIpc.HeapSys)
	fmt.Printf("Heap Memory Idle: %s\n", resIpc.HeapIdle)
	fmt.Printf("Heap Memory In Use: %s\n", resIpc.HeapInuse)
	fmt.Printf("Heap Memory Released: %s\n", resIpc.HeapReleased)
	fmt.Printf("Heap Objects: %d\n\n", resIpc.HeapObjects)

	fmt.Printf("Bootstrap Stack Usage: %s\n", resIpc.StackInuse)
	fmt.Printf("Stack Memory Obtained: %s\n", resIpc.StackSys)
	fmt.Printf("MSpan Structures Usage: %s\n", resIpc.MSpanInuse)
	fmt.Printf("MSpan Structures Obtained: %s\n", resIpc.MSpanSys)
	fmt.Printf("MCache Structures Usage: %s\n", resIpc.MCacheInuse)
	fmt.Printf("MCache Structures Obtained: %s\n", resIpc.MCacheSys)
	fmt.Printf("Profiling Bucket Hash Table OBtained: %s\n", resIpc.BuckHashSys)
	fmt.Printf("GC Metadata Obtained: %s\n", resIpc.GCSys)
	fmt.Printf("Other System Allocation Obtained: %s\n\n", resIpc.OtherSys)

	fmt.Printf("Next GC Recycle: %s\n", resIpc.NextGC)
	fmt.Printf("Since Last GC Time: %s\n", resIpc.LastGC)
	fmt.Printf("Total GC Pause: %s\n", resIpc.PauseTotalNs)
	fmt.Printf("Last GC Pause: %s\n", resIpc.PauseNs)
	fmt.Printf("GC Times: %d\n", resIpc.NumGC)

	return err
}