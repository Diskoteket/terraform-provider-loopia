terraform {
  required_providers {
    loopia = {
      source = "hashicorp.com/edu/loopia"
    }
  }
}

provider "loopia" {}

resource "loopia_subdomain" "something_example_com" {
  subdomain = "something"
  domain    = "example.com"
}
