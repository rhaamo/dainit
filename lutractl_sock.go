package main

import (
	"syscall"
	"os"
	"bytes"
	"io"
	"fmt"
	"strings"
)

func socketInitctl() {
	err := syscall.Mkfifo("/run/initctl", 0600)
	if err != nil {
		println("[lutra][socket] Mkfifo error:", err)
	}

	for {
		f, err := os.OpenFile("/run/initctl", os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			println("[lutra][socket] OpenFile error:", err)
		}

		var buff bytes.Buffer
		io.Copy(&buff, f)
		go handleMessage(buff.String())
		f.Close()
	}
}

// We may support that one day, for now, just a placeholder
func handleMessage(msg string) {
	fmt.Printf("[lutra][socket] Got query: '%s'\n", msg)

	switch strings.TrimSpace(msg) {
		/* Runlevels 0 to 6 */
	case "i   0", "i    1":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   1":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   2":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   3":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   4":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   5":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   6":
		println("[lutra][socket] Unsupported telinit thing")

		/* Tell init to switch to single user mode */
	case "i   S":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   s":
		println("[lutra][socket] Unsupported telinit thing")

		/* Tell init to re-examine /etc/inittab */
	case "i   Q":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   q":
		println("[lutra][socket] Unsupported telinit thing")

		/* A B C process only thoses in /etc/inittab */
	case "i   A":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   a":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   B":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   b":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   C":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   c":
		println("[lutra][socket] Unsupported telinit thing")

		/* Tell init to re-execute itself */
	case "i   U":
		println("[lutra][socket] Unsupported telinit thing")
	case "i   u":
		println("[lutra][socket] Unsupported telinit thing")

	default:
		println("[lutra][socket] unknown telinit thing")
	}
}