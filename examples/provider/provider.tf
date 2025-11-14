terraform {
  required_providers {
    loopia = {
      source = "diskoteket/loopia"
    }
  }
}

provider "loopia" {
  username = "my-api-user-name@loopiaapi"
  password = "my-api-user-password"
}
