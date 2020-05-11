package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/logging"

	"github.com/bendrucker/terraform-cloud-cli/cmd"
)

func main() {
	cli := cmd.NewCLI()

	logging.SetOutput()

	status, err := cli.Run()
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(status)
}
