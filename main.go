package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/kassett/terraform-provider-balena/balena"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := plugin.ServeOpts{
		ProviderFunc: balena.Provider,
		Debug:        debugMode,
		ProviderAddr: "registry.terraform.io/kassett/balena",
	}

	plugin.Serve(&opts)
}
