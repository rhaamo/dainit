package main

import (
	"os"
	"github.com/valyala/gorpc"
	"sync"
	"github.com/urfave/cli"
	"strings"
	"github.com/olekukonko/tablewriter"
	toposort "github.com/philopon/go-toposort"
	"github.com/go-clog/clog"
	"fmt"
)

var (
	// LutraVersion should match the one in lutractl/main.go
	LutraVersion = "0.1"

	// StartupServices Should only be used for the FIRST startup
	// StartupServices is the in-memory map list of processes started on a full-start boot
	StartupServices = make(map[ServiceType][]*StartupService)

	// LoadedServices is used for any other actions, start, stop, etc.
	LoadedServices 		= make(map[ServiceName]*Service)
	LoadedServicesMu	= sync.RWMutex{}

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
	app.Commands = []cli.Command {
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

var CmdServicesTree = cli.Command {
	Name: "services-tree",
	Usage: "List the services tree",
	Description: "List the services tree",
	Action: dumpServicesTree,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "confdir", Value: "/etc/lutrainit", Usage: "Lutrainit config directory"},
	},
}

func dumpServicesTree(ctx *cli.Context) error {
	err := setupLogging(false); if err != nil {
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

	println("len services", len(LoadedServices))
	graph := toposort.NewGraph(len(LoadedServices))

	// Add nodes
	for _, v := range LoadedServices {
		ok := graph.AddNode(string(v.Name)); if ok {
			fmt.Printf("Added node '%s'\n", v.Name)
		} else {
			fmt.Printf("Cannot add node '%s'\n", v.Name)
		}
	}

	// Add edges
	for _, s := range LoadedServices {
		// WantedBy
		if s.WantedBy != "" {
			dep := string(LoadedServices[ServiceName(s.WantedBy)].Name)
			ok := graph.AddEdge(string(s.Name), dep); if ok {
				fmt.Printf("Added edge from '%s' to '%s'\n", s.Name, dep)
			} else {
				fmt.Printf("Cannot add edge from '%s' to '%s'\n", s.Name, dep)
			}
		}

		// After
		for _, aft := range s.After {
			dep := string(LoadedServices[ServiceName(aft)].Name)
			ok := graph.AddEdge(string(s.Name), dep); if ok {
				fmt.Printf("Added edge from '%s' to '%s'\n", s.Name, dep)
			} else {
				fmt.Printf("Cannot add edge from '%s' to '%s'\n", s.Name, dep)
			}
		}
		// Before
		for _, bf := range s.Before {
			dep := string(LoadedServices[ServiceName(bf)].Name)
			ok := graph.AddEdge(string(s.Name), dep); if ok {
				fmt.Printf("Added edge from '%s' to '%s'\n", s.Name, dep)
			} else {
				fmt.Printf("Cannot add edge from '%s' to '%s'\n", s.Name, dep)
			}
		}
		// Requires
		for _, req := range s.Requires {
			dep := string(LoadedServices[ServiceName(req)].Name)
			ok := graph.AddEdge(string(s.Name), dep); if ok {
				fmt.Printf("Added edge from '%s' to '%s'\n", s.Name, dep)
			} else {
				fmt.Printf("Cannot add edge from '%s' to '%s'\n", s.Name, dep)
			}
		}
	}

	// sort
	list, ok := graph.Toposort()
	if !ok {
		clog.Error(2, "Cycle detected :(")
		return fmt.Errorf("Cycle detected")
	}

	fmt.Printf("\nBoot services order:\n")
	for _, s := range list {
		if strings.HasSuffix(s,".target") || strings.HasSuffix(s, ".state"){
			fmt.Printf("+ %s\n", s)
		} else {
			fmt.Printf(" \\__ %s\n", s)
		}
	}

	return nil
}

var CmdServicesList = cli.Command {
	Name: "services-list",
	Usage: "List the services list",
	Description: "List the services",
	Action: dumpServicesList,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "confdir", Value: "/etc/lutrainit", Usage: "Lutrainit config directory"},
	},
}

func dumpServicesList(ctx *cli.Context) error {
	err := setupLogging(false); if err != nil {
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
