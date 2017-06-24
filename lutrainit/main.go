// dainit is a simple init system for Linux intended for non-server uses
// (ie. for a programmer's home Linux system.)
//
// It aims to be easy to use.
package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"github.com/go-clog/clog"
	"syscall"
	"fmt"
	"github.com/valyala/gorpc"
	"encoding/gob"
	"bytes"
)

var (
	// LutraVersion should match the one in lutractl/main.go
	LutraVersion = "0.1"

	// StartupServices Should only be used for the FIRST startup
	// StartupServices is the in-memory map list of processes started on a full-start boot
	StartupServices = make(map[ServiceType][]*StartupService)

	// LoadedServices is used for any other actions, start, stop, etc.
	LoadedServices = make(map[ServiceName]*Service)

	// NetFs design the list of known network file systems to be avoided mounted at boot
	NetFs = []string{"nfs", "nfs4", "smbfs", "cifs", "codafs", "ncpfs", "shfs", "fuse", "fuseblk", "glusterfs", "davfs", "fuse.glusterfs"}
	// VirtFs design the list of known virtual file systems to avoid unmounting at shutdown
	VirtFs = []string{"proc", "sysfs", "tmpfs", "devtmpfs", "devpts"}

	// GoRPCServer for client
	GoRPCServer = &gorpc.Server{}

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

// Runs a command, setting up Stdin/Stdout/Stderr to be the standard OS
// ones, and not /dev/null. This will block until the command finishes.
func run(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	return c.Run()
}

// Runs a command and waits for it to finish.
func runquiet(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	return c.Run()
}

// SetHostname set the hostname
func SetHostname(hostname []byte) {
	proc, err := os.Create("/proc/sys/kernel/hostname")
	if err != nil {
		clog.Error(2, err.Error())
		return
	}
	defer proc.Close()

	_, err = proc.Write(hostname)
	if err != nil {
		clog.Error(2, err.Error())
		return
	}
}

func main() {
	reexec := os.Getenv("LUTRAINIT_REEXECING")
	MainConfig.StartedReexec = reexec == "true"

	err := setupLogging(false); if err != nil {
		println("[lutra] Error: This is going bad, could not setup logging", err.Error())
		// we have no choice
		// PANIC PANIC PANIC
		os.Exit(-1)
	}

	if MainConfig.StartedReexec {
		fmt.Println("Re-exec-ing of lutrainit in progress")
		err := gobelinFromFile()
		if err != nil {
			println("error deserializing structs from files. expect misbehaviors or panics.")
		}
	}

	// First of first, who are we ?
	clog.Info("~~ LutraInit %s starting...", LutraVersion)
	clog.Info("~~ Build commit %s", LutraBuildGitHash)
	clog.Info("~~ Build time %s", LutraBuildTime)

	if !thePidOne() {
		clog.Warn("[lutra] I'm sorry but I'm supposed to be run as an init.")
		os.Exit(-1)
	}

	// First of all, we need to be sure we have a correct PATH setted
	// This is usefull if we use lutrainit in an initramfs since PATH would be unset
	curEnvPath := os.Getenv("PATH")
	if len(strings.TrimSpace(curEnvPath)) == 0 {
		os.Setenv("PATH", "/usr/local/sbin:/sbin:/bin:/usr/sbin:/usr/bin")
		clog.Info("[lutra] Empty $PATH, fixed.")
	}
	clog.Info("[lutra] Current $PATH is: %s", os.Getenv("PATH"))

	if !MainConfig.StartedReexec {
		// Remount root as rw.
		//
		// This should be handled by mount -a according to mount(8), since the flags that
		// / is mounted with don't match /etc/fstab, but for some reason it's not remounting
		// root as rw even though it's not ro in /etc/fstab. TODO: Look into this..
		clog.Info("[lutra] Remounting root filesystem")
		Remount("/")
	}

	// Start socket in background
	go socketInitctl()

	if !MainConfig.StartedReexec {
		// Mount local filesytems
		clog.Info("[lutra] Mounting local file systems")
		MountAllExcept(NetFs)

		// Activate swap partitions, mount -a doesn't do this since they aren't really mounted anywhere
		run("swapon", "-a")

		// Set the hostname for getty to be happy.
		if hostname, err := ioutil.ReadFile("/etc/hostname"); err == nil {
			SetHostname(hostname)
		} else {
			clog.Error(2, "[lutra] Error setting hostname: %s", err.Error())
		}

		// There's a little (dare I say, a lot?) of black magic that seems to
		// happen on a modern Linux systems for filesystems.
		//
		// Time was, everything would go into /etc/fstab.
		//
		// If you read "Demystifying the init system" (https://felipec.wordpress.com/2013/11/04/init/)
		// he manually mounts /proc, /sys, /run, etc.
		//
		// When I do that on my Debian system, I get an "already mounted" error
		// despite it not being in /etc/fstab, so I don't mount them here (either
		// the kernel is doing it, or Debian is lying about init being the first
		// process.)
		//
		// However, /dev/shm is neither automounted, nor in fstab, and without it
		// Chromium won't start, so it's done manually here.
		//
		// (It should probably just go into my fstab to be honest, but then it seems
		// that I'd need to manually add it every time I install a new system..)
		Mount("tmpfs", "shm", "/dev/shm", "mode=1777,nosuid,nodev")
	}

	// Parse configurations, reexec is counted as reloading
	ReloadConfig(MainConfig.StartedReexec, false)

	// We finally have a filesystem mounted and the configuration is parsed
	if err := setupLogging(true); err != nil {
		clog.Error(2, "Failed to add file logging to logger: %s", err.Error())
	}

	if !MainConfig.StartedReexec {
		// Start all services from StartupServices in the right Needs/Provide order
		StartServices()
	}

	// Kick Zombies out in the background
	go reapChildren()

	ManageGettys()

	// The ttys exited. Kill processes, unmount filesystems and halt the system.
	doShutdown(false)
}

func thePidOne() bool {
	if os.Getpid() == 1 {
		return true
	}

	return false
}

func setupLogging(withFile bool) (err error) {
	err = clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level: clog.TRACE, // record all logs
		BufferSize: 100,     // log async, 0 is sync
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

func gobelinToFile() (err error) {
	bls := new(bytes.Buffer)
	bgl := new(bytes.Buffer)

	ls := gob.NewEncoder(bls)
	gl := gob.NewEncoder(bgl)

	err = ls.Encode(LoadedServices)
	if err != nil {
		clog.Error(2, "error encoding LoadedServices: %s", err.Error())
		return err
	}

	err = gl.Encode(GettysList)
	if err != nil {
		clog.Error(2, "error encoding GettysList: %s", err.Error())
		return err
	}

	err = ioutil.WriteFile(lsFnameSerialized, bls.Bytes(), 0644)
	if err != nil {
		clog.Error(2, "error saving serialized file %s: %s", lsFnameSerialized, err.Error())
		return err
	}

	err = ioutil.WriteFile(glFnameSerialized, bgl.Bytes(), 0644)
	if err != nil {
		clog.Error(2, "error saving serialized file %s: %s", glFnameSerialized, err.Error())
		return err
	}

	clog.Info("LoadedServices and GettysList have been serialized to file")
	return nil
}

func gobelinFromFile() (err error) {
	lsBytes, err := ioutil.ReadFile(lsFnameSerialized)
	if err != nil {
		clog.Error(2, "error loading serialized file %s: %s", lsFnameSerialized, err.Error())
		return err
	}

	glBytes, err := ioutil.ReadFile(glFnameSerialized)
	if err != nil {
		clog.Error(2, "error loading serialized file %s: %s", glFnameSerialized, err.Error())
		return err
	}

	ls := gob.NewDecoder(bytes.NewReader(lsBytes))
	if err != nil {
		clog.Error(2, "error decoding LoadedServices: %s", err.Error())
		return err
	}

	err = ls.Decode(&LoadedServices)
	if err != nil {
		clog.Error(2, "error mapping decoded LoadedServices: %s", err.Error())
		return err
	}

	gl := gob.NewDecoder(bytes.NewReader(glBytes))
	if err != nil {
		clog.Error(2, "error decoding GettysList: %s", err.Error())
		return err
	}

	err = gl.Decode(&GettysList)
	if err != nil {
		clog.Error(2, "error mapping decoded GettysList: %s", err.Error())
		return err
	}

	// Remove old serializing files
	if err = os.Remove(lsFnameSerialized); err != nil {
		clog.Error(2, "error removing file %s: %s", lsFnameSerialized, err.Error())
		return err
	}
	if err = os.Remove(glFnameSerialized); err != nil {
		clog.Error(2, "error removing file %s: %s", glFnameSerialized, err.Error())
		return err
	}

	clog.Info("LoadedServices and GettysList have been de-serialized from file")

	return nil
}

// ReExecInit take care of stopping RPC server, removing file logger and serialization before execve()
func ReExecInit() {
	fmt.Println("reexecing...")

	// Stop GoRPC
	GoRPCServer.Stop()

	// Remove file logger
	setupLogging(false)

	// Serialize the LoadedServices struct
	err := gobelinToFile()
	if err != nil {
		clog.Error(2, "error serializing structures. expect misbehaviors or panics.")
	}

	// Prepare new environment
	os.Setenv("LUTRAINIT_REEXECING", "true")

	if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
		fmt.Println("reexec failed:", err)
		// What to do ?
		// Exit ? Continue ? Panic ?
	}
}
