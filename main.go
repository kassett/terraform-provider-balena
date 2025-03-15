package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/kassett/terraform-provider-balena/balena"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: balena.Provider})
}
