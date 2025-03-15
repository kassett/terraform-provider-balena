
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