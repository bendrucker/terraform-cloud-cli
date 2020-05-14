# This file was generated by GoReleaser. DO NOT EDIT.
class TerraformCloudCli < Formula
  desc "Interactive CLI for operating workspaces in Terraform Cloud"
  homepage "https://github.com/bendrucker/terraform-cloud-cli"
  version "0.1.3"
  bottle :unneeded

  if OS.mac?
    url "https://github.com/bendrucker/terraform-cloud-cli/releases/download/v0.1.3/terraform-cloud-cli_0.1.3_darwin_amd64.tar.gz"
    sha256 "acc70a25087138e28e4ae9878e15f1ccf7516107f93e2efcb568f96fa52a5dcd"
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/bendrucker/terraform-cloud-cli/releases/download/v0.1.3/terraform-cloud-cli_0.1.3_linux_amd64.tar.gz"
      sha256 "a0d3d3731e6bd444ae19e9247414e64f9355972e370785da79a9adf57ef1edc5"
    end
  end

  def install
    bin.install "terraform-cloud"
  end
end
