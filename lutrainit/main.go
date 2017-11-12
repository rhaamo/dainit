package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"github.com/valyala/gorpc"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	// LutraVersion should match the one in lutractl/main.go
	LutraVersion = "0.1"

	// StartupServices Should only be used for the FIRST startup
	// StartupServices is the in-memory map list of processes started on a full-start boot
	StartupServices = make(map[ServiceName][]ServiceName)
	// StartupTargets ordered slice
	StartupTargets = make([]ServiceName, 0)

	// LoadedServices is used for any other actions, start, stop, etc.
	LoadedServices = make(map[ServiceName]*Service)
	// LoadedServicesMu tex to avoid issues
	LoadedServicesMu = sync.RWMutex{}

	// NetFs design the list of known network file systems to be avoided mounted at boot
	NetFs = []string{"nfs", "nfs4", "smbfs", "cifs", "codafs", "ncpfs", "shfs", "fuse", "fuseblk", "glusterfs", "davfs", "fuse.glusterfs"}
	// VirtFs design the list of known virtual file systems to avoid unmounting at shutdown
	VirtFs = []string{"proc", "sysfs", "tmpfs", "devtmpfs", "devpts"}

	// GoRPCServer for client
	GoRPCServer = &gorpc.Server{}
	// GoRPCStarted or not
	GoRPCStarted = false

	//ShuttingDown is used to break various check loops like in getty
	ShuttingDown bool

	lsFnameSerialized = "/run/lutrainit.reexec.ls.bin"
	glFnameSerialized = "/run/lutrainit.reexec.gl.bin"

	// Theses two last should only filled by LDFLAGS, see Makefile

	// LutraBuildTime is the time of the build
	LutraBuildTime string
	// LutraBuildGitHash is the git sha1 of the commit based on
	LutraBuildGitHash string
)

func main() {
	app := cli.NewApp()
	app.Name = "lutrainit"
	app.Usage = "lutra init daemon"
	app.Version = LutraVersion
	app.Commands = []cli.Command{
		CmdServicesTree,
		CmdServicesList,
		CmdSysinit,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)

	// No argument will start the system init processing
	if len(os.Args) <= 1 {
		os.Args = []string{"lutrainit", "sysinit"}
	}

	app.Run(os.Args)
}

// CmdServicesTree cli command
var CmdServicesTree = cli.Command{
	Name:        "services-tree",
	Usage:       "List the services tree",
	Description: "List the services tree",
	Action:      dumpServicesTree,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "confdir", Value: "/etc/lutrainit", Usage: "Lutrainit config directory"},
	},
}

func dumpServicesTree(ctx *cli.Context) error {
	err := setupLogging(false)
	if err != nil {
		println("[lutra] Error: This is going bad, could not setup logging", err.Error())
		// we have no choice
		// PANIC PANIC PANIC
		os.Exit(-1)
	}

	var baseDir string

	if !ctx.IsSet("confdir") {
		baseDir = "/etc/lutrainit"
	} else {
		baseDir = ctx.String("confdir")
	}
	if err = ReloadConfig(false, baseDir, false); err != nil {
		return err
	}

	time.Sleep(500 * time.Microsecond)

	// Sort the services
	SortServicesForBoot()

	// Print the tree
	for idx, target := range StartupTargets {
		fmt.Printf("+ [%d] %s\n", idx, target)
		for idx, service := range StartupServices[target] {
			fmt.Printf(" - [%d] %s\n", idx, service)
		}
	}

	return nil
}

// CmdServicesList cli command
var CmdServicesList = cli.Command{
	Name:        "services-list",
	Usage:       "List the services list",
	Description: "List the services",
	Action:      dumpServicesList,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "confdir", Value: "/etc/lutrainit", Usage: "Lutrainit config directory"},
	},
}

func dumpServicesList(ctx *cli.Context) error {
	err := setupLogging(false)
	if err != nil {
		println("[lutra] Error: This is going bad, could not setup logging", err.Error())
		// we have no choice
		// PANIC PANIC PANIC
		os.Exit(-1)
	}

	var baseDir string

	if !ctx.IsSet("confdir") {
		baseDir = "/etc/lutrainit"
	} else {
		baseDir = ctx.String("confdir")
	}
	ReloadConfig(false, baseDir, false)

	time.Sleep(500 * time.Microsecond)

	data := [][]string{}
	for _, service := range LoadedServices {
		if !service.IsService() {
			continue
		}
		data = append(data, []string{
			string(service.Name),
			string(service.WantedBy),
			service.Type,
			strings.Join(service.Requires, ","),
			strings.Join(service.After, ","),
			strings.Join(service.Before, ","),
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"name", "target", "type", "requires", "after", "before"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	return nil
}
