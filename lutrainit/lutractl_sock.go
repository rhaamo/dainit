package main

import (
	"fmt"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"github.com/rhaamo/lutrainit/lutrainit/tools"
	"github.com/valyala/gorpc"
	"time"
	"runtime"
)

var (
	startTime = time.Now()
)

func socketInitctl() {
	d := gorpc.NewDispatcher()

	// Returns the init daemon version
	d.AddFunc("version", func() *ipc.Version {
		return &ipc.Version{
			ServerVersion: LutraVersion,
			ServerBuildHash: LutraBuildGitHash,
			ServerBuildTime: LutraBuildTime,
		}
	})

	// Returns the daemon system stats
	d.AddFunc("stats", func() *ipc.SysStatus {
		return returnStats()
	})

	// Returns processes statuses
	d.AddFunc("status", func(status *ipc.AskStatus) map[ipc.ServiceName]*ipc.LoadedService {
		return returnStatus(status)
	})

	s := gorpc.NewUnixServer("/run/ottersock", d.NewHandlerFunc())
	if err := s.Serve(); err != nil {
		println("[lutra][socket] Starting GoRPC error", err)
		return
	}
	println("[lutra][socket] GoRPC started", s.Addr)
	defer s.Stop()
}

func returnStats() *ipc.SysStatus {
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	return &ipc.SysStatus{
		Uptime: tools.TimeSincePro(startTime),

		NumGoroutine: runtime.NumGoroutine(),

		MemAllocated: tools.FileSize(int64(m.Alloc)),
		MemTotal: tools.FileSize(int64(m.TotalAlloc)),
		MemSys: tools.FileSize(int64(m.Sys)),
		Lookups: m.Lookups,
		MemMallocs: m.Mallocs,
		MemFrees: m.Frees,

		HeapAlloc: tools.FileSize(int64(m.HeapAlloc)),
		HeapSys: tools.FileSize(int64(m.HeapSys)),
		HeapIdle: tools.FileSize(int64(m.HeapIdle)),
		HeapInuse: tools.FileSize(int64(m.HeapInuse)),
		HeapReleased: tools.FileSize(int64(m.HeapReleased)),
		HeapObjects: m.HeapObjects,

		StackInuse: tools.FileSize(int64(m.StackInuse)),
		StackSys: tools.FileSize(int64(m.StackSys)),
		MSpanInuse: tools.FileSize(int64(m.MSpanInuse)),
		MSpanSys: tools.FileSize(int64(m.MSpanSys)),
		MCacheInuse: tools.FileSize(int64(m.MCacheInuse)),
		MCacheSys: tools.FileSize(int64(m.MCacheSys)),
		BuckHashSys: tools.FileSize(int64(m.BuckHashSys)),
		GCSys: tools.FileSize(int64(m.GCSys)),
		OtherSys: tools.FileSize(int64(m.OtherSys)),

		NextGC: tools.FileSize(int64(m.NextGC)),

		LastGC: fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(m.LastGC))/1000/1000/1000),
		PauseTotalNs: fmt.Sprintf("%.1fs", float64(m.PauseTotalNs)/1000/1000/1000),
		PauseNs: fmt.Sprintf("%.3fs", float64(m.PauseNs[(m.NumGC+255)%256])/1000/1000/1000),
		NumGC: m.NumGC,
	}
}

func returnStatus(req *ipc.AskStatus) map[ipc.ServiceName]*ipc.LoadedService {
	if req.All {
		return LoadedServices
	}

	if proc, exists := LoadedServices[ipc.ServiceName(req.Name)]; exists {
		procList := make(map[ipc.ServiceName]*ipc.LoadedService)
		procList[ipc.ServiceName(req.Name)] = proc
		return procList
	}
	return nil
}