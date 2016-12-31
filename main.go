// dainit is a simple init system for Linux.
// 
// dainit will currently do the following:
//  1. Set the hostname
//  2. Remount the root filesystem rw
//  3. Start udevd to probe for devices (TODO: Find a systemd-free replacement, but udevd is
//     what my laptop has for now...)
//  4. Start a tty and wait for a login.
//  5. Kill running processes, unmount filesystems, and poweroff the system once that login
//     session ends.
package main

import (
	"fmt"
	"io/ioutil"
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
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	defer proc.Close()

	n, err := proc.Write([]byte(hostname))
	if n != len(hostname) || err != nil {
		fmt.Fprintf(os.Stderr, "Hostname not set incorrectly (%d != %d, %v) \n", n, len(hostname), err)
	}
}

func main() {
	// Set the hostname for getty to be happy.
	if hostname, err := ioutil.ReadFile("/etc/hostname"); err == nil {
		SetHostname(string(hostname))
	}

	// Remount root as rw
	println("Remounting root filesystem")
	Remount("/")

	// Mount local filesytems
	println("Mounting local file systems")
	netfs := []string{"nfs", "nfs4", "smbfs", "cifs", "codafs", "ncpfs", "shfs", "fuse", "fuseblk", "glusterfs", "davfs", "fuse.glusterfs"}
	virtfs := []string{"proc", "sysfs", "tmpfs", "devtmpfs", "devpts"}
	MountAll(netfs)

	// TODO: Find or write a systemd-free replacement for udevd and get rid of this.
	systemdHacks()

	// Launch a get tty process.
	// If we don't Setsid, we'll get an "inappropriate ioctl for device"
	// error upon starting the login shell.
	cmd := exec.Command("agetty", "--noclear", "tty1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	cmd.Run()

	// The tty exited. Kill processes, unmount filesystems and halt the system.
	println("Stopping udevd...")
	run("udevadm", "control", "--exit")
	println("Killing the world...")
	KillAll()
	println("unmounting filesystems...")
	UnmountAll(append(netfs, virtfs...))

	// Halt the system.
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}
