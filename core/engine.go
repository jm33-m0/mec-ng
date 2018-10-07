package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jm33-m0/mec-ng/utils"
	"gopkg.in/cheggaaa/pb.v1"
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
					utils.PrintCyan("[*] custom args: %s", strings.Join(TailArgs, " "))
				}
			}
		}

		// cd to work dir
		err := os.Chdir(Environ.WorkDir)
		if err != nil {
			utils.PrintRed("[-] cannot enter target directory: %s", err.Error())
			return
		}
		utils.PrintCyan("[*] working under: %s", Environ.WorkDir)
		utils.PrintCyan("[*] target: %s", Environ.TargetList)
	}

	t := time.Now()
	Environ.TimeStamp = t.Format("20060102150405")
}

// Dispatcher : read cmdline args and do the job
func Dispatcher() {
	switch Mode {
	case "custom":
		if Module == "" {
			utils.PrintRed("[-] please specify the executable to run")
			return
		}
		run(Module)
	case "zoomeye":
		fmt.Printf("[*] please run %s/built-in/zoomeye/zoomeye.py manually\n", Environ.MecRoot)
	case "masscan":
		masscan(MasscanRange)
	case "xmir":
		xmir(MasscanXML, Filter)
	}
}

func run(mod string) {
	// get abs path of executable
	modArray := strings.Split(mod, "/")
	modArray = modArray[1:]
	mod = Environ.MecRoot + "/" + strings.Join(modArray, "/")

	utils.PrintCyan("[*] started %s with %d workers", mod, JobCnt)

	lines, err := utils.FileToLines(Environ.TargetList)
	if err != nil {
		utils.PrintError("[-] unable to open target list: %s, %s", Environ.TargetList, err.Error())
		return
	}

	var wg sync.WaitGroup
	i := 1 // job counter

	// start a progress bar
	length, err := utils.GetFileLength(Environ.TargetList)
	if err != nil {
		utils.PrintSuccess("[-] Error getting file length: %s", err.Error())
		return
	}
	bar := pb.StartNew(length)
	bar.SetRefreshRate(50 * time.Millisecond)

	for _, line := range lines {
		ip := strings.Trim(line, "\n")

		go func() {
			wg.Add(1)
			defer wg.Done()
			argsArray := append(TailArgs, ip)
			args := strings.Join(argsArray, " ")

			// utils.PrintCyan("working on %s", ip)
			if err := utils.ExecCmd(mod, args); err != nil {
				utils.PrintError("[-] Error on %s: %s", ip, err.Error())
			}
			bar.Increment()

		}()

		i++

		if i == JobCnt && &wg != nil {
			i = 0
			wg.Wait()
		}

	}

	// for {
	// 	time.Sleep(1 * time.Second)
	// 	// TODO check if any process is still running, if none found, tell the routine to exit
	// }
}

func masscan(rangelist string) {
	// use masscan to grab a list of targets
	utils.PrintCyan("[*] starting masscan")
	utils.PrintCyan("[*] please be patient, masscan might take some time if target list is large")

	prog := "masscan"
	args := fmt.Sprintf("-iL %s -c %s/conf/masscan.conf -oX %s", rangelist, Environ.MecRoot, Environ.MecRoot+"/output/"+Environ.TimeStamp+"-masscan.xml")

	utils.ExecCmd(prog, args)
}

func xmir(xml string, filter string) {
	// parse masscan xml result
	utils.PrintCyan("[*] xmir started")
	outfile := Environ.MecRoot + "/output/" + Environ.TimeStamp + ".xmirlist"

	// if ip list doesn't exist, parse the XML file to get one
	if _, err := os.Stat(outfile); os.IsNotExist(err) {
		utils.PrintCyan("parsing masscan result...")
		utils.XML2List(xml, outfile, filter)
	}
}
