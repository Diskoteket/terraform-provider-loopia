terraform {
  required_providers {
    loopia = {
      source = "diskoteket/loopia"
    }
  }
}

provider "loopia" {}

data "loopia_domain" "example_com" {
  name = "example.com"
}

output "example_com_expiration_date" {
  value = data.loopia_domain.example_com.expiration_date
}