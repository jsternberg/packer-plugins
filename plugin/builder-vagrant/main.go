package main

import (
	"github.com/jsternberg/packer-plugins/builder/vagrant"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(vagrant.Builder))
	server.Serve()
}
