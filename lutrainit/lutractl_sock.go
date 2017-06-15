package main

import (
	"fmt"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"github.com/valyala/gorpc"
)

func socketInitctl() {
	gorpc.RegisterType(&ipc.IpcVersion{})

	s := gorpc.NewUnixServer("/run/ottersock", handleMessage)
	if err := s.Serve(); err != nil {
		println("[lutra][socket] Starting GoRPC error", err)
		return
	}
	println("[lutra][socket] GoRPC started", s.Addr)

	defer s.Stop()
}

func handleMessage(clientAddr string, request interface{}) interface{} {
	fmt.Printf("[lutra][socket] Got message: '%s'\n", request)

	switch request {
	case "version":
		return returnVersion()
	default:
		fmt.Printf("[lutra][socket] Got unknown RPC '%s'", request)
	}

	return request
}

func returnVersion() interface{} {
	return &ipc.IpcVersion{
		ServerVersion: LutraVersion,
		ServerBuildHash: LutraBuildGitHash,
		ServerBuildTime: LutraBuildTime,
	}
}