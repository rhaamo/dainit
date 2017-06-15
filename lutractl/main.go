package main

import (
	"github.com/urfave/cli"
	"os"
)

var (
	LutraVersion = "0.1"
	// Theses two last should only filled by LDFLAGS, see Makefile
	LutraBuildTime string
	LutraBuildGitHash string
)



func main() {
	app := cli.NewApp()
	app.Name = "lutractl_ctl"
	app.Usage = "lutra init control client"
	app.Version = LutraVersion
	app.Commands = []cli.Command {
		CmdVersion,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}