
provider "balena" {}

terraform {
  required_providers {
    balena = {
      source = "registry.terraform.io/kassett/balena"
    }
  }
}

data "balena_device" "this" {
  uuid = "d8262387f955b30572d79a6fd2fa78b4"
}

data "balena_services" "this" {
  fleet_id = data.balena_device.this.fleet_id
}

data "balena_service_variables" "this" {
  service_id = data.balena_services.this.services[0].service_id
}

data "balena_device_tags" "this" {
  device_uuid = data.balena_device.this.uuid
}

output "device_tags" {
  value = data.balena_device_tags.this.tags
}