package main

import (
	"fmt"
	"os"

	"github.com/keithalucas/nvme-go/pkg/nvme"
)

func register() {
	deviceFile, err := nvme.RegisterDevice("test", "nqn.2021-12.io.longhorn.volume:test", "10.0.0.59", 4420)

	if err != nil {
		fmt.Printf("error %v\n", err)
	}

	fmt.Printf("%s\n", deviceFile)
}

func unregister() {
	nvme.UnregisterDevice("test")
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "register" {
		register()
	} else if os.Args[1] == "unregister" {
		unregister()
	} else {
		fmt.Printf("unknown argument")
	}
}
