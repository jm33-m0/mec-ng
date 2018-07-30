package main

import (
	"github.com/jm33-m0/mec-ng/core"
)

func main() {
	// init
	core.PrintBanner()
	core.ArgParse()

	// env
	core.Config()

	// do the job
	core.Dispatcher()
}
