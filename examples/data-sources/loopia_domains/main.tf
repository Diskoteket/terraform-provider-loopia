terraform {
  required_providers {
    loopia = {
      source = "diskoteket/loopia"
    }
  }
}

provider "loopia" {}

data "loopia_domains" "all" {}

output "all_domains" {
  value = data.loopia_domains.all.domains
}
