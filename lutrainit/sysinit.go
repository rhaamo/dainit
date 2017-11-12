package main

import (
	"github.com/go-clog/clog"
	"github.com/urfave/cli"
)

// CmdSysinit cli command
var CmdSysinit = cli.Command {
	Name: "sysinit",
	Usage: "Sysinit",
	Description: "Sysinit",
	Action: sysinit,
	Flags: []cli.Flag{},
}

func setupLogging(withFile bool) (err error) {
	err = clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level:      clog.TRACE, // record all logs
		BufferSize: 100,        // log async, 0 is sync
	})
	if err != nil {
		println("Whoops, cannot initialize logging to console:", err.Error())
		return err
	}

	if withFile {
		err = clog.New(clog.FILE, clog.FileConfig{
			Level:      clog.TRACE,
			BufferSize: MainConfig.Log.BufferLen,
			Filename:   MainConfig.Log.Filename,
			FileRotationConfig: clog.FileRotationConfig{
				Rotate:   MainConfig.Log.Rotate,
				Daily:    MainConfig.Log.Daily,
				MaxSize:  1 << uint(MainConfig.Log.MaxSize),
				MaxLines: MainConfig.Log.MaxLines,
				MaxDays:  MainConfig.Log.MaxDays,
			},
		})
		if err != nil {
			clog.Error(2, "Cannot initialize log to file: %s", err.Error())
		}
	}
	return err
}
