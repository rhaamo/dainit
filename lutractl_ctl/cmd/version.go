package cmd

import (
	"github.com/urfave/cli"
	"net"
	"fmt"
	"bufio"
	"strings"
	"github.com/rhaamo/lutrainit/lutractl"
)

var Cmd = cli.Command {
	Name: "version",
	Usage: "Shows init version",
	Description: "Shows init version",
	Action: getVersion,
	Flags: []cli.Flag{},
}

func getVersion(ctx *cli.Context) error {
	conn, err := net.Dial("unix", "/run/ottersock")
	if err != nil {
		println("lutractl dial socket error:", err.Error())
	}
	fmt.Fprintf(conn, "version\r\n")
	respVers, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		println("lutractl reading response error:", err.Error())
	}

	respSplit := strings.Split(respVers, ";")
	fmt.Printf("Client version: %s\nRunning init: %s\nBuilt on: %s\nCommit sha: %s\n",
		lutractl.LutraVersion, respSplit[0], respSplit[1],respSplit[2])

	return nil
}