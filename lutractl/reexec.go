package main

import (
	"github.com/urfave/cli"
	"errors"
	"fmt"
)

// CmdReload CLI object
var CmdReexec = cli.Command {
	Name: "reexec",
	Usage: "Re-exec init daemon",
	Description: "Re-exec init daemon",
	Action: doReexec,
	Flags: []cli.Flag{},
}

func doReexec(ctx *cli.Context) error {
	if !IsRoot() {
		return errors.New("only root can do that")
	}

	_, err := GorpcDispatcherClient.Call("reexec", nil)

	if err != nil {
		return err
	}

	fmt.Printf("Init may re-exec, check logs.")

	return err
}