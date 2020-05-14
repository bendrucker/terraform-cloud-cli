terraform {
  backend "remote" {
    organization = "org"

    workspaces {
      prefix = "ws-"
    }
  }
}
