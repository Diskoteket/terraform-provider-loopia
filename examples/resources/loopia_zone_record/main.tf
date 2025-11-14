terraform {
  required_providers {
    loopia = {
      source = "diskoteket/loopia"
    }
  }
}

provider "loopia" {}

resource "loopia_subdomain" "something_example_com" {
  subdomain = "something"
  domain    = "example.com"
}

resource "loopia_zone_record" "something_example_com" {
  domain    = resource.loopia_subdomain.something_example_com.domain
  subdomain = resource.loopia_subdomain.something_example_com.subdomain
  record = {
    type  = "A"
    value = "192.0.2.1"
  }
}

output "record_id" {
  value = loopia_zone_record.something_example_com.record
}

