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
	MecRoot    string
	TargetList string
	WorkDir    string
	TimeStamp  string
	Config     string
}

// Environ : init env vars
var Environ Env

// Config : read env vars and configs
func Config() {
	exec, _ := os.Executable()
	Environ.MecRoot = filepath.Dir(exec)
	Environ.Config = Environ.MecRoot + "/conf/mec.conf"

	if Module != "" {
		// workdir
		workpath := strings.Split(Module, "/")
		workpath = workpath[1 : len(workpath)-1]
		workdir := "/" + strings.Join(workpath, "/")
		Environ.WorkDir = Environ.MecRoot + workdir

		// target list
		listFile := strings.Split(IPList, "/")
		listFile = listFile[1:]
		list := "/" + strings.Join(listFile, "/")
		Environ.TargetList = Environ.MecRoot + list

		// custom_args
		if lines, err := utils.FileToLines(Environ.Config); err == nil {
			for _, line := range lines {
				line = strings.Trim(line, "\n")
				if strings.HasPrefix(line, "custom_args") {
					lineArray := strings.Split(line, "=")
					TailArgs = strings.Split(lineArray[1], " ")
					log.Print("[*] custom args: ", TailArgs)
				}
			}
		}

		// cd to work dir
		err := os.Chdir(Environ.WorkDir)
		if err != nil {
			log.Fatal("[-] cannot enter target directory: ", err)
		}
		log.Println("[*] working under: ", Environ.WorkDir)
		log.Println("[*] target: ", Environ.TargetList)
	}

	t := time.Now()
	Environ.TimeStamp = t.Format("20060102150405")
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
		log.Println("[*] starting zoomeye.py")
		prog := "python3"
		args := fmt.Sprintf("%s/built-in/zoomeye.py", Environ.MecRoot)
		utils.ExecCmd(prog, args)
	case "masscan":
		masscan()
	case "xmir":
		xmir(MasscanXML, Filter)
	}
}

func run(mod string) {
	// get abs path of executable
	modArray := strings.Split(mod, "/")
	modArray = modArray[1:]
	mod = Environ.MecRoot + "/" + strings.Join(modArray, "/")

	log.Printf("[*] started %s with %d workers", mod, JobCnt)

	lines, err := utils.FileToLines(Environ.TargetList)
	if err != nil {
		log.Printf("[-] unable to open target list: %s", Environ.TargetList)
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
	log.Println("[*] starting masscan")

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
		log.Println("parsing masscan result...")
		utils.XML2List(xml, outfile, filter)
	}
}
