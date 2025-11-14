terraform {
  required_providers {
    loopia = {
      source = "diskoteket/loopia"
    }
  }
}

provider "loopia" {}

data "loopia_zonerecords" "something_example_com" {
  subdomain = "something"
  domain    = "example.com"
}

output "something_example_com_zone_records" {
  value = data.loopia_zone_records.something_example_com.zone_records
}
