
provider "balena" {}

terraform {
  required_providers {
    balena = {
      source = "registry.terraform.io/kassett/balena"
    }
  }
}

data "balena_fleet" "this" {
  slug = "org/fleet-name"
}

data "balena_fleet_variables" "this" {
  fleet_id = data.balena_fleet.this.fleet_id
}

data "balena_device" "this" {
  uuid = "a-randomly-generated-uuid"
}

data "balena_fleet" "fleet_for_device" {
  fleet_id = data.balena_device.this.fleet_id
}

data "balena_sensitive_fleet_variable" "this" {
  fleet_id = data.balena_fleet.this.fleet_id
  variable_name = "VARIABLE_NAME_FOR_DEVICE"
}
data "balena_services" "this" {
  fleet_id = data.balena_fleet.this.fleet_id
}

data "balena_service_variables" "this" {
  service_id = data.balena_services.this.services[0].service_id
}

data "balena_service_variable" "this" {
  service_id = data.balena_services.this.services[0].service_id
  variable_name = "DEVICE_ENVIRONMENT"
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

output "balena_fleet_variables" {
  value = data.balena_fleet_variables.this.variables
}

output "balena_service_variables" {
  value = data.balena_service_variables.this.variables
}

output "balena_fleet_variable" {
  value = data.balena_sensitive_fleet_variable.this.value
  sensitive = true
}

output "balena_service_variable" {
  value = data.balena_service_variable.this.value
}

output "balena_variables_for_fleet" {
  value = data.balena_fleet.fleet_for_device
}