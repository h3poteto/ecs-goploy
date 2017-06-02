package main

import (
	"fmt"
	"os"

	"github.com/crowdworks/ecs-goploy/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
