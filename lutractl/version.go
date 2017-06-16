package main

import (
	"github.com/urfave/cli"
	"github.com/rhaamo/lutrainit/shared/ipc"
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
	res, err := GorpcDispatcherClient.Call("version", nil)

	resIpc := res.(*ipc.IpcVersion)

	fmt.Printf("Client version: %s\nRunning init: %s\nBuilt on: %s\nCommit sha: %s\n",
		LutraVersion, resIpc.ServerVersion, resIpc.ServerBuildTime, resIpc.ServerBuildHash)
	return err
}