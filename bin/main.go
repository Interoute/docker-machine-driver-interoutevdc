package main

import (
        "github.com/radu-stefanache/docker-machine-driver-interoutevdc"
        "github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
        plugin.RegisterDriver(interoutevdc.NewDriver("", ""))
}
