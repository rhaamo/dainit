package main

import (
	"fmt"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"github.com/rhaamo/lutrainit/lutrainit/tools"
	"github.com/valyala/gorpc"
	"time"
	"runtime"
	"github.com/go-clog/clog"
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

	d.AddFunc("shutdown", func() {
		clog.Info("[lutra] I was asked to shutdown, goodbye!")
		doShutdown(false)
		// will never return, sorry
	})

	d.AddFunc("reboot", func() {
		clog.Info("[lutra] I was asked to reboot, seeya!")
		doShutdown(true)
		// will never return, sorry
	})

	// Returns processes statuses
	d.AddFunc("status", func(status *ipc.AskStatus) map[ipc.ServiceName]*ipc.Service {
		return returnStatus(status)
	})

	d.AddFunc("reload", func() *ipc.AnswerReload {
		err := ReloadConfig(true, true)
		if err != nil {
			return &ipc.AnswerReload{Err: true, ErrStr: err.Error()}
		}
		return &ipc.AnswerReload{Err: false}
	})

	d.AddFunc("start", func(req *ipc.ServiceAction) *ipc.ServiceActionAnswer {
		answer := &ipc.ServiceActionAnswer{Name: req.Name, Action: ipc.Start}

		if proc, exists := LoadedServices[ServiceName(req.Name)]; exists {
			err := CheckAndStartService(proc)
			if err != nil {
				answer.Err = true
				answer.ErrStr = err.Error()
				return answer
			}
		}

		return answer
	})

	d.AddFunc("stop", func(req *ipc.ServiceAction) *ipc.ServiceActionAnswer {
		answer := &ipc.ServiceActionAnswer{Name: req.Name, Action: ipc.Stop}

		if proc, exists := LoadedServices[ServiceName(req.Name)]; exists {
			err := CheckAndStopService(proc)
			if err != nil {
				answer.Err = true
				answer.ErrStr = err.Error()
				return answer
			}
		}

		return answer
	})

	d.AddFunc("reexec", func() {
		go ReExecInit()
		return
	})

	GoRPCServer = gorpc.NewUnixServer("/run/ottersock", d.NewHandlerFunc())
	clog.Info("[lutra] RPC starting on socket: %s", GoRPCServer.Addr)
	GoRPCStarted = true

	if err := GoRPCServer.Serve(); err != nil {
		GoRPCStarted = false
		clog.Error(2, "[lutra][socket] Starting GoRPC error", err.Error())
		return
	}
	GoRPCStarted = false
	clog.Info("GoRPC stopped")
	//defer GoRPCServer.Stop()
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

func returnStatus(req *ipc.AskStatus) (services map[ipc.ServiceName]*ipc.Service) {
	services = make(map[ipc.ServiceName]*ipc.Service)

	if req.All {
		for k, v := range LoadedServices {
			services[ipc.ServiceName(k)] = &ipc.Service{
				Name: ipc.ServiceName(v.Name),
				Type: v.Type,
				Description: v.Description,
				State: ipc.RunState(v.State),
				LastAction: ipc.LastAction(v.LastAction),
				LastActionAt: v.LastActionAt,
				LastMessage: v.LastMessage,
				Deleted: v.Deleted,
			}
		}
	} else {
		if proc, exists := LoadedServices[ServiceName(req.Name)]; exists {
			services[ipc.ServiceName(proc.Name)] = &ipc.Service{
				Name:         ipc.ServiceName(proc.Name),
				Type:         proc.Type,
				Description:  proc.Description,
				State:        ipc.RunState(proc.State),
				LastAction:   ipc.LastAction(proc.LastAction),
				LastActionAt: proc.LastActionAt,
				LastMessage:  proc.LastMessage,
				Deleted: proc.Deleted,
			}
		} else {
			return nil
		}
	}

	return services
}