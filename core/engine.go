package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jm33-m0/mec-ng/utils"
)

// Env : env vars for mec
type Env struct {
	MecRoot   string
	WorkDir   string
	TimeStamp string
}

// Environ : init env vars
var Environ Env

// Config : read env vars and configs
func Config(mod string) {
	exec, _ := os.Executable()
	Environ.MecRoot = filepath.Dir(exec)
	Environ.WorkDir = Environ.MecRoot + "/modules/" + mod
	Environ.TimeStamp = time.Now().Format("20110504111515")
}

// Dispatcher : read cmdline args and do the job
func Dispatcher() {
	switch Mode {
	case "custom":
		if Module == "" {
			log.Fatal("[-] please specify the executable to run")
		}
		run(Module)
	case "zoomeye":
		log.Println("[*] Starting zoomeye.py")
		prog := "python3"
		args := fmt.Sprintf("%s/scripts/zoomeye.py", Environ.MecRoot)
		utils.ExecCmd(prog, args)
	case "masscan":
		masscan()
	case "xmir":
		xmir(MasscanXML, Filter)
	}
}

func run(mod string) {
	log.Printf("[*] Started %s with %d workers", mod, JobCnt)

	lines, err := utils.FileToLines(IPList)
	if err != nil {
		log.Printf("[-] Unable to open target list: %s", IPList)
		log.Print(err)
		return
	}

	var wg sync.WaitGroup
	i := 0 // job counter
	for _, line := range lines {
		ip := strings.Trim(line, "\n")
		go func() {
			wg.Add(1)
			defer wg.Done()
			argsArray := append(TailArgs, ip)
			args := strings.Join(argsArray, ",")
			utils.ExecCmd(mod, args)
		}()
		i++
		if i == JobCnt && &wg != nil {
			i = 0
			wg.Wait()
		}
	}
	for {
		time.Sleep(1 * time.Second)
		// TODO check if any process is still running, if none found, tell the routine to exit
	}
}

func masscan() {
	// use masscan to grab a list of targets
	log.Println("[*] Starting masscan")

	prog := "masscan"
	args := fmt.Sprintf("-c %s/conf/masscan.conf -oX %s", Environ.MecRoot, Environ.MecRoot+"/output/"+Environ.TimeStamp+"-masscan.xml")

	utils.ExecCmd(prog, args)
}

func xmir(xml string, filter string) {
	// parse masscan xml result
	log.Println("[*] xmir started")
	outfile := Environ.MecRoot + "/output/" + Environ.TimeStamp + ".xmirlist"

	// if ip list doesn't exist, parse the XML file to get one
	if _, err := os.Stat(outfile); os.IsNotExist(err) {
		log.Println("Parsing masscan result...")
		utils.XML2List(xml, outfile, filter)
	}
}
