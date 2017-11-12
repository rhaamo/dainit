package main

import (
	"errors"
	"github.com/urfave/cli"
)

// CmdShutdown CLI object
var CmdShutdown = cli.Command{
	Aliases:     []string{"halt"},
	Name:        "shutdown",
	Usage:       "Shutdowns the system",
	Description: "Shutdowns the system",
	Action:      doShutdown,
	Flags:       []cli.Flag{},
}

// CmdReboot CLI object
var CmdReboot = cli.Command{
	Aliases:     []string{"restart"},
	Name:        "reboot",
	Usage:       "Reboot the system",
	Description: "Reboot the system",
	Action:      doReboot,
	Flags:       []cli.Flag{},
}

func doShutdown(ctx *cli.Context) error {
	if !IsRoot() {
		return errors.New("only root can do that")
	}

	_, err := GorpcDispatcherClient.Call("shutdown", nil)
	return err
}

func doReboot(ctx *cli.Context) error {
	if !IsRoot() {
		return errors.New("only root can do that")
	}

	_, err := GorpcDispatcherClient.Call("reboot", nil)
	return err
}
