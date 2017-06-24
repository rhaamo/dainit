package main

import (
	"github.com/urfave/cli"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"fmt"
	"time"
)

// CmdStatus CLI object
var CmdStatus = cli.Command {
	Name: "status",
	Usage: "Shows init processes status",
	Description: "Shows init processes status",
	Action: getStatus,
	Flags: []cli.Flag{},
}

func getStatus(ctx *cli.Context) error {

	req := &ipc.AskStatus{}

	if ctx.Args().Present() {
		req.Name = ctx.Args().First()
		req.All = false
	} else {
		req.All = true
	}

	res, err := GorpcDispatcherClient.Call("status", req)

	resIpc := res.(map[ipc.ServiceName]*ipc.Service)

	if len(resIpc) == 0 && !req.All {
		fmt.Printf("No service matching '%s'\n", req.Name)
		return nil
	} else if len(resIpc) == 0 && req.All {
		fmt.Printf("No statuses returned at all, Houston, we've a problem.\n")
		return nil
	}

	for _, loadedService := range resIpc {
		fmt.Printf("Service: %s, of type %s\n", loadedService.Name, loadedService.Type)
		fmt.Printf("Description: %s\n", loadedService.Description)
		if loadedService.Deleted {
			fmt.Printf("WARNING: This service init have been deleted from configuration directory.\n")
		}
		fmt.Printf("Status: %s\n", loadedService.State.String())
		lastActionAt := time.Unix(loadedService.LastActionAt, 0).Format(time.RFC1123Z)
		fmt.Printf("Last action: %s at %s\n", loadedService.LastAction.String(), lastActionAt)
		fmt.Printf("Last message: %s\n\n", loadedService.LastMessage)
	}

	return err
}