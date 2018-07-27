package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Env : env vars for mec
type Env struct {
	MecRoot string
	WorkDir string
}

// Environ : init env vars
var Environ Env

// Config : read env vars and configs
func Config(mod string) {
	exec, _ := os.Executable()
	Environ.MecRoot = filepath.Dir(exec)
	Environ.WorkDir = Environ.MecRoot + "/modules/" + mod
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
		ExecCmd(prog, args)
	case "masscan":
		masscan()
	}
}

func run(mod string) {
	log.Printf("[*] Started %s with %d workers", mod, JobCnt)

	lines, err := FileToLines(IPList)
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
			ExecCmd(mod, args)
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
	args := fmt.Sprintf("-c %s/conf/masscan.conf", Environ.MecRoot)

	ExecCmd(prog, args)
}
