package main

import (
	"github.com/urfave/cli"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"fmt"
	"time"
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

	resIpc := res.(map[ipc.ServiceName]*ipc.IpcLoadedService)

	for _, loadedService := range resIpc {
		fmt.Printf("Service: %s, of type %s\n", loadedService.Name, loadedService.Type)
		fmt.Printf("Description: %s\n", loadedService.Description)
		fmt.Printf("Status: %s\n", loadedService.State.String())
		if loadedService.Type == "simple" && loadedService.State == ipc.Started && loadedService.LastKnownPID >= 2{
			fmt.Printf("Lask known PID: %d\n", loadedService.LastKnownPID)
		}
		lastActionAt := time.Unix(loadedService.LastActionAt, 0).Format(time.RFC1123Z)
		fmt.Printf("Last action: %s at %s\n", loadedService.LastAction.String(), lastActionAt)
		fmt.Printf("Last message: %s\n\n", loadedService.LastMessage)
	}

	return err
}