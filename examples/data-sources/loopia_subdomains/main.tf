terraform {
  required_providers {
    loopia = {
      source = "diskoteket/loopia"
    }
  }
}

provider "loopia" {}

data "loopia_subdomains" "example_com" {
  domain = "example.com"
}

output "example_com_subdomains" {
  value = data.loopia_subdomains.example_com.subdomains
}
