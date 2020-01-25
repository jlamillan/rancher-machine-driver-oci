package main

import (
	"github.com/jlamillan/docker-machine-driver-oci/pkg/drivers/oci"
	"github.com/rancher/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(oci.NewDriver("", ""))
}

