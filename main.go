package main

import (
	"context"
	"log"

	"github.com/awlsring/terraform-provider-proxmox/proxmox"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
func main() {
	err := providerserver.Serve(context.Background(), proxmox.New, providerserver.ServeOpts{
		Address: "tmp.terraform.io/awlsring/proxmox",
	})

	if err != nil {
		log.Fatal(err.Error())
	}
}
