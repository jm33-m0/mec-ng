package core

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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
		if len(TailArgs) == 0 {
			log.Fatal("[-] please specify the executable to run")
		}
		run(TailArgs[0])
	case "zoomeye":
		log.Println("[*] Starting zoomeye.py")
		cmdstr := fmt.Sprintf("python3 %s/scripts/zoomeye.py", Environ.MecRoot)
		cmd := exec.Command(cmdstr)
		cmd.Run()
	case "masscan":
		masscan()
	}
}

func run(mod string) {
	log.Printf("[*] Started %s with %d workers", mod, JobCnt)

	lines, err := FileToLines(IPList)
	if err != nil {
		log.Printf("[-] Unable to open %s", IPList)
		return
	}

	i := 0 // job counter
	for _, line := range lines {
		ip := strings.Trim(line, "\n")
		var wg sync.WaitGroup
		go func() {
			wg.Add(1)
			toExec := append(TailArgs, ip)
			cmd := exec.Command(strings.Join(toExec, ","))
			err = cmd.Run()
			if err != nil {
				log.Print("[-] Error running task: ", err)
				return
			}
			wg.Done()
		}()
		i++
		if i == JobCnt && &wg != nil {
			i = 0
			wg.Wait()
		}
	}
}

func masscan() {
	// use masscan to grab a list of targets
	log.Println("[*] Starting masscan")

	args := fmt.Sprintf("-c %s/conf/masscan.conf", Environ.MecRoot)
	cmd := exec.Command("masscan", strings.Split(args, " ")...)

	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	errScanner := bufio.NewScanner(stderr)
	outScanner := bufio.NewScanner(stdout)
	// scanner.Split(bufio.ScanLines)
	go func() {
		for outScanner.Scan() {
			m := outScanner.Text()
			fmt.Println(m)
		}
	}()
	go func() {
		for errScanner.Scan() {
			e := errScanner.Text()
			fmt.Println(e)
		}
	}()
	cmd.Wait()
}
