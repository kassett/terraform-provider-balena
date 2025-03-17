
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

data "balena_services" "this" {
  fleet_id = data.balena_fleet.this.fleet_id
}

output "fleet_attributes" {
  value = data.balena_fleet.this
}

output "device_attributes" {
  value = data.balena_device.this
}

output "balena_services_attributes" {
  value = data.balena_services.this
}