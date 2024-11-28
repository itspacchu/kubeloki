package main

import (
	"github.com/itspacchu/kubeloki/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
