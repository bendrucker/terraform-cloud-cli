package main

import (
	"fmt"
	"os"

	"github.com/bendrucker/terraform-cloud-cli/cmd"
)

func main() {
	cli := cmd.NewCLI()

	status, err := cli.Run()
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(status)
}
