package main

import (
	"fmt"
	"github.com/go-clog/clog"
	"github.com/gyuho/goraph"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"github.com/valyala/gorpc"
	"os"
	"strings"
	"sync"
	"time"
	//"github.com/davecgh/go-spew/spew"
)

var (
	// LutraVersion should match the one in lutractl/main.go
	LutraVersion = "0.1"

	// StartupServices Should only be used for the FIRST startup
	// StartupServices is the in-memory map list of processes started on a full-start boot
	StartupServices = make(map[ServiceType][]*StartupService)

	// LoadedServices is used for any other actions, start, stop, etc.
	LoadedServices   = make(map[ServiceName]*Service)
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

	graph := goraph.NewGraph()

	// Add nodes
	for _, v := range LoadedServices {
		node := goraph.NewNode(string(v.Name))
		v.Node = node.ID()
		ok := graph.AddNode(node)
		if ok {
			fmt.Printf("Added node '%s'\n", v.Name)
		} else {
			fmt.Printf("Cannot add node '%s': node already exists\n", v.Name)
		}
	}

	// Add edges
	for _, s := range LoadedServices {
		// WantedBy
		if s.WantedBy != "" {
			err = graph.AddEdge(LoadedServices[ServiceName(s.WantedBy)].Node, s.Node, 100)
			if err == nil {
				fmt.Printf("Added WantedBy edge from '%s' to '%s'\n", s.WantedBy, s.Name)
			} else {
				fmt.Printf("Cannot add WantedBy edge from '%s' to '%s': %s\n", s.WantedBy, s.Name, err)
			}
		}

		// After
		for _, aft := range s.After {
			err = graph.AddEdge(LoadedServices[ServiceName(aft)].Node, s.Node, 100)
			if err == nil {
				fmt.Printf("Added After edge from '%s' to '%s'\n", aft, s.Name)
			} else {
				fmt.Printf("Cannot add After edge from '%s' to '%s': %s\n", aft, s.Name, err)
			}
		}
		// Before
		for _, bf := range s.Before {
			err = graph.AddEdge(s.Node, LoadedServices[ServiceName(bf)].Node, 100)
			if err == nil {
				fmt.Printf("Added Before edge from '%s' to '%s'\n", s.Name, bf)
			} else {
				fmt.Printf("Cannot add Before edge from '%s' to '%s': %s\n", s.Name, bf, err)
			}
		}
		// Requires
		for _, req := range s.Requires {
			err := graph.AddEdge(LoadedServices[ServiceName(req)].Node, s.Node, 100)
			if err == nil {
				fmt.Printf("Added Require edge from '%s' to '%s'\n", req, s.Name)
			} else {
				fmt.Printf("Cannot add Require edge from '%s' to '%s'\n", req, s.Name)
			}
		}
	}

	// sort
	list, ok := goraph.TopologicalSort(graph)
	if !ok {
		clog.Error(2, "Cycle detected :(")
		return fmt.Errorf("Cycle detected")
	}

	fmt.Printf("\nBoot services order:\n")
	for _, s := range list {
		if strings.HasSuffix(s.String(), ".target") || strings.HasSuffix(s.String(), ".state") {
			fmt.Printf("+ %s\n", s)
		} else {
			fmt.Printf(" - %s\n", s)
		}
	}

	return nil
}

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

	data := [][]string{}
	for _, service := range LoadedServices {
		data = append(data, []string{
			string(service.Name),
			service.Type,
			strings.Join(service.Requires, ","),
			strings.Join(service.After, ","),
			strings.Join(service.Before, ","),
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"name", "type", "requires", "after", "before"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	return nil
}
