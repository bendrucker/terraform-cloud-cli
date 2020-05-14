terraform {
  backend "remote" {
    organization = "org"

    workspaces {
      name = "ws-name"
    }
  }
}
