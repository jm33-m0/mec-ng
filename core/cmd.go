package core

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"strings"
)

// LogoEncoded : base64 encoded logo, ascii art
const LogoEncoded = "ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgXyBfXyBfX18gICBfX18gIF9fXyAgICAgIF8gX18gICBfXyBfIAogfCAnXyBgIF8gXCAvIF8gXC8gX198X19fX3wgJ18gXCAvIF9gIHwKIHwgfCB8IHwgfCB8ICBfXy8gKF98X19fX198IHwgfCB8IChffCB8CiB8X3wgfF98IHxffFxfX198XF9fX3wgICAgfF98IHxffFxfXywgfAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICB8X19fLyAK"

var (
	// IPList : target list
	IPList string
	// Mode : working mode
	Mode string
	// ThdCnt : how many tasks per time
	ThdCnt int
	// Module : which module to use
	Module string
	// TailArgs : more args in the tail
	TailArgs []string
)

// ArgParse : parse cmd line args for package core
func ArgParse() {
	IPList = *flag.String("iplist", "", "target ip list")
	Mode = *flag.String("mode", "", "working mode")
	ThdCnt = *flag.Int("thd", 100, "how many tasks per time")
	Module = *flag.String("module", "", "which module to use")

	flag.Parse()

	TailArgs = flag.Args()
}

// PrintBanner : print mec-ng ascii logo
func PrintBanner() {
	logo, err := base64.StdEncoding.DecodeString(LogoEncoded)
	if err != nil {
		log.Panicf("Logo error: ", err)
	}
	fmt.Println(string(logo))
	fmt.Println(strings.Repeat(" ", 26) + "by jm33-ng")
}
