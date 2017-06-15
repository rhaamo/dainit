package main

import (
	"github.com/urfave/cli"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"github.com/valyala/gorpc"
	"fmt"
)

var CmdVersion = cli.Command {
	Name: "version",
	Usage: "Shows init version",
	Description: "Shows init version",
	Action: getVersion,
	Flags: []cli.Flag{},
}

func getVersion(ctx *cli.Context) error {
	gorpc.RegisterType(&ipc.IpcVersion{})

	c := gorpc.NewUnixClient("/run/ottersock")
	c.Start()
	defer c.Stop()

	res, err := c.Call("version")

	resIpc := res.(*ipc.IpcVersion)

	fmt.Printf("Client version: %s\nRunning init: %s\nBuilt on: %s\nCommit sha: %s\n",
		LutraVersion, resIpc.ServerVersion, resIpc.ServerBuildTime, resIpc.ServerBuildHash)
	return err
}