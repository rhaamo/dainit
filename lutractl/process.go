package main

import (
	"errors"
	"fmt"
	"dev.sigpipe.me/dashie/lutrainit/shared/ipc"
	"github.com/urfave/cli"
)

// CmdStart CLI object
var CmdStart = cli.Command{
	Name:        "start",
	Usage:       "Start process",
	Description: "Start process",
	Action:      doStart,
	Flags:       []cli.Flag{},
}

// CmdStop CLI object
var CmdStop = cli.Command{
	Name:        "stop",
	Usage:       "Stop process",
	Description: "Stop process",
	Action:      doStop,
	Flags:       []cli.Flag{},
}

// CmdRestart CLI obkect
var CmdRestart = cli.Command{
	Name:        "restart",
	Usage:       "Restart process",
	Description: "Restart process",
	Action:      doRestart,
	Flags:       []cli.Flag{},
}

func doStart(ctx *cli.Context) error {
	if !IsRoot() {
		return errors.New("only root can do that")
	}

	if !ctx.Args().Present() {
		return cli.NewExitError("process name required", -1)
	}

	procName := ctx.Args().First()

	res, err := GorpcDispatcherClient.Call("start", &ipc.ServiceAction{Name: procName, Action: ipc.Start})
	if err != nil {
		return err
	}

	resIpc := res.(*ipc.ServiceActionAnswer)

	if resIpc.Err {
		fmt.Printf("Error starting %s: %s\n", resIpc.Name, resIpc.ErrStr)
		return errors.New(resIpc.ErrStr)
	}

	fmt.Printf("Service %s started.\n", resIpc.Name)

	return nil
}

func doStop(ctx *cli.Context) error {
	if !IsRoot() {
		return errors.New("only root can do that")
	}

	if !ctx.Args().Present() {
		return cli.NewExitError("process name required", -1)
	}

	procName := ctx.Args().First()

	res, err := GorpcDispatcherClient.Call("stop", &ipc.ServiceAction{Name: procName, Action: ipc.Stop})
	if err != nil {
		return err
	}

	resIpc := res.(*ipc.ServiceActionAnswer)

	if resIpc.Err {
		fmt.Printf("Error stopping %s: %s\n", resIpc.Name, resIpc.ErrStr)
		return errors.New(resIpc.ErrStr)
	}

	fmt.Printf("Service %s stopped.\n", resIpc.Name)

	return nil
}

func doRestart(ctx *cli.Context) error {
	err := doStop(ctx)
	if err != nil {
		return err
	}
	err = doStart(ctx)
	if err != nil {
		return err
	}
	return nil
}
