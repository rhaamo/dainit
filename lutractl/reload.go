package main

import (
	"errors"
	"dev.sigpipe.me/dashie/lutrainit/shared/ipc"
	"github.com/urfave/cli"
)

// CmdReload CLI object
var CmdReload = cli.Command{
	Name:        "reload",
	Usage:       "Reload init configs",
	Description: "Reload init configs",
	Action:      doReload,
	Flags:       []cli.Flag{},
}

func doReload(ctx *cli.Context) error {
	res, err := GorpcDispatcherClient.Call("reload", nil)

	if err != nil {
		return err
	} else if res == nil {
		return errors.New("result is <nil>")
	}

	resIpc := res.(*ipc.AnswerReload)

	if resIpc.Err {
		println("Reload error:", resIpc.ErrStr)
	} else {
		println("Reload successful.")
	}

	return err
}
