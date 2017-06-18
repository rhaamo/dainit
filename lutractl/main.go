package main

import (
	"github.com/urfave/cli"
	"os"
	"github.com/valyala/gorpc"
	"github.com/rhaamo/lutrainit/shared/ipc"
)

var (
	LutraVersion = "0.1"
	// Theses two last should only filled by LDFLAGS, see Makefile
	LutraBuildTime string
	LutraBuildGitHash string

	// Dispatcher to dispatch things
	GorpcDispatcher *gorpc.Dispatcher
	GorpcDispatcherClient	*gorpc.DispatcherClient
	GorpcClient		*gorpc.Client
)



func main() {
	app := cli.NewApp()
	app.Name = "lutractl"
	app.Usage = "lutra init control client"
	app.Version = LutraVersion
	app.Commands = []cli.Command {
		CmdVersion,
		CmdStats,
		CmdStatus,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)

	// Main RPC initialization
	gorpc.RegisterType(&ipc.IpcSysStatus{})
	gorpc.RegisterType(&ipc.IpcVersion{})
	gorpc.RegisterType(&ipc.IpcAskStatus{})

	GorpcDispatcher = gorpc.NewDispatcher()

	GorpcDispatcher.AddFunc("status", func(status *ipc.IpcAskStatus) map[ipc.ServiceName]*ipc.IpcLoadedService {
		println("wanting client status")
		return nil
	})

	GorpcClient = gorpc.NewUnixClient("/run/ottersock")
	GorpcClient.Start()
	defer GorpcClient.Stop()

	GorpcDispatcherClient = GorpcDispatcher.NewFuncClient(GorpcClient)


	// Let's go baby
	app.Run(os.Args)
}
