// dainit is a simple init system for Linux intended for non-server uses
// (ie. for a programmer's home Linux system.)
//
// It aims to be easy to use.
package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
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

// Set's the hostname for the kernel.
func SetHostname(hostname string) {
	proc, err := os.Create("/proc/sys/kernel/hostname")
	if err != nil {
		log.Println(err)
	}
	defer proc.Close()

	n, err := proc.Write([]byte(hostname))
	if n != len(hostname) || err != nil {
		log.Println(err)
	}
}

func main() {
	// Remount root as rw.
	//
	// This should be handled by mount -a according to mount(8), since the flags that
	// / is mounted with don't match /etc/fstab, but for some reason it's not remounting
	// root as rw even though it's not ro in /etc/fstab. TODO: Look into this..
	println("Remounting root filesystem")
	Remount("/")

	// Mount local filesytems
	println("Mounting local file systems")
	netfs := []string{"nfs", "nfs4", "smbfs", "cifs", "codafs", "ncpfs", "shfs", "fuse", "fuseblk", "glusterfs", "davfs", "fuse.glusterfs"}
	virtfs := []string{"proc", "sysfs", "tmpfs", "devtmpfs", "devpts"}

	MountAllExcept(netfs)
	// Activate swap partitions, mount -a doesn't do this since they aren't really mounted anywhere
	run("swapon", "-a")

	// Set the hostname for getty to be happy.
	if hostname, err := ioutil.ReadFile("/etc/hostname"); err == nil {
		SetHostname(string(hostname))
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

	// Parse all the configs in /etc/dainit. Finally!
	services, err := ParseServiceConfigs("/etc/dainit")
	if err != nil {
		log.Println(err)
	}

	StartServices(services)

	go reapChildren()

	// Launch an appropriate number of getty processes on ttys.
	if conf, err := os.Open("/etc/dainit.conf"); err != nil {
		// If the config doesn't exist or can't be opened, use the defaults.
		Gettys(nil, false)
	} else {
		if autologins, persist, err := ParseSetupConfig(conf); err != nil {
			// We don't want to defer this because otherwise it won't
			// get executed until the system is shutting down..
			conf.Close()

			// If the config couldn't be parsed, used the defaults
			log.Println(err)
			Gettys(nil, false)
		} else {
			conf.Close()
			Gettys(autologins, persist)
		}
	}

	// The tty exited. Kill processes, unmount filesystems and halt the system.
	// TODO: Run shutdown scripts for services that are started instead
	// of just sending them a SIGTERM right off the bat..
	println("Killing everything I can find...")
	KillAll()

	// This needs to be done after all the processes are dead, otherwise
	// it will fail due to being in use.
	println("Unmounting filesystems...")
	UnmountAllExcept(append(netfs, virtfs...))

	// Halt the system explicitly to prevent a kernel panic.
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}
