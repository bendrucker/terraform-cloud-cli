terraform {
  backend "remote" {
    hostname = "host.name"
    organization = "org"

    workspaces {
      name = "ws-name"
    }
  }
}