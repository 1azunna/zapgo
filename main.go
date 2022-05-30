package main

import (
	"fmt"

	"github.com/1azunna/zapgo/docker/cmd"
)

var version string
var date string

func main() {
	fmt.Println("Version:\t", version)
	fmt.Println("Build Time:\t", date)
	cmd.Execute()
}
