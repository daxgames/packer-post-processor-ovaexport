package main

import (
	"github.com/mitchellh/packer/packer/plugin"
  "github.com/daxgames/packer-post-processor-ovfexport/post-processor"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterPostProcessor(new(ovfexport.PostProcessor))
	server.Serve()
}
