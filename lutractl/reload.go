package main

import (
	"github.com/urfave/cli"
	"github.com/rhaamo/lutrainit/shared/ipc"
)

// CmdReload CLI object
var CmdReload = cli.Command {
	Name: "reload",
	Usage: "Reload init configs",
	Description: "Reload init configs",
	Action: doReload,
	Flags: []cli.Flag{},
}

func doReload(ctx *cli.Context) error {
	res, err := GorpcDispatcherClient.Call("reload", nil)

	resIpc := res.(*ipc.AnswerReload)

	if resIpc.Err {
		println("Reload error:", resIpc.ErrStr)
	} else {
		println("Reload successfull.")
	}

	return err
}