
provider "balena" {}

terraform {
  required_providers {
    balena = {
      source = "registry.terraform.io/kassett/balena"
    }
  }
}

data "balena_fleet" "this" {
  slug = "tagup/stage-router"
}

data "balena_device" "this" {
  uuid = "6ac8cf056c432579b2c8fd62a183e264"
}

output "device_name" {
  value = data.balena_device.this.uuid
}