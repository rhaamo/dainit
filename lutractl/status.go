package main

import (
	"github.com/urfave/cli"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"fmt"
)

var CmdStatus = cli.Command {
	Name: "status",
	Usage: "Shows init processes status",
	Description: "Shows init processes status",
	Action: getStatus,
	Flags: []cli.Flag{},
}

// TODO: ability to get only one process status

func getStatus(ctx *cli.Context) error {
	req := &ipc.IpcAskStatus{All: true}

	// Ask status for all processes
	res, err := GorpcDispatcherClient.Call("status", req)

	fmt.Printf("%+v\n", res)

	//resIpc := res.(map[ipc.IpcServiceType]*ipc.IpcProcess)

	//for k, v := range resIpc {
	//	fmt.Printf("%s / %s: %d", k, v.Name, v.RunState)
	//}

	return err
}