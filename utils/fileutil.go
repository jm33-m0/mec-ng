package utils

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

// Address : host>address
type Address struct {
	Addr     string `xml:"addr,attr"`
	Addrtype string `xml:"addrtype,attr"`
}

// State : host>ports>port>state
type State struct {
	State     string `xml:"state,attr"`
	Reason    string `xml:"reason,attr"`
	ReasonTTL string `xml:"reason_ttl,attr"`
}

// Service : host>ports>port>service
type Service struct {
	Name   string `xml:"name,attr"`
	Banner string `xml:"banner,attr"`
}

// Ports : host>ports
type Ports []struct {
	Protocol string `xml:"protocol,attr"`
	Portid   string `xml:"portid,attr"`

	State   State   `xml:"state"`
	Service Service `xml:"service"`
}

// Host : host field in XML
type Host struct {
	XMLName xml.Name `xml:"host"`
	Endtime string   `xml:"endtime,attr"`

	Address Address `xml:"address"`
	Ports   Ports   `xml:"ports>port"`
}

// XML2List : Parse masscan result, pick useful items and save them to a list file
func XML2List(xmlfile string, outfile string, filter string) {

	xmlStream, err := os.Open(xmlfile)
	if err != nil {
		PrintError("Failed to open XML file: %s", err)
		return
	}
	defer func() {
		if err = xmlStream.Close(); err != nil {
			PrintError(err.Error())
		}
	}()

	// open outfile
	outf, err := OpenFileStream(outfile)
	if err != nil {
		log.Println("Error opening ", outfile+"\n", err)
		return
	}
	defer func() {
		err = CloseFileStream(outf)
		if err != nil {
			PrintError(err.Error())
		}
	}()

	decoder := xml.NewDecoder(xmlStream)
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "host" {
				var h Host
				err = decoder.DecodeElement(&h, &se)
				if err != nil {
					PrintError(err.Error())
				}

				// since mostly we have just one port to detect
				address := h.Address.Addr
				// port := h.Ports[0].Portid

				banner := ""
				if len(h.Ports) > 0 {
					banner = h.Ports[0].Service.Banner
				}

				// write desired host to file
				if searchHost(filter, banner) {
					err = AppendToFile(outf, address)
					if err != nil {
						PrintError(err.Error())
					}
				}
			}
		default:
		}
	}
}

// AppendToFile : append a line to target file
func AppendToFile(file *os.File, line string) (err error) {
	// write appendly
	if _, err = file.Write([]byte(line + "\n")); err != nil {
		PrintError("Write err: %s\nwriting line: %s", err.Error(), line)
		return err
	}
	return nil
}

// OpenFileStream : open file for writing
func OpenFileStream(filepath string) (file *os.File, err error) {
	// open outfile
	file, err = os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		PrintError("%s : Failed to open file\n%s", filepath, err.Error())
		return nil, err
	}
	return file, nil
}

// CloseFileStream : Close file when we are done with it
func CloseFileStream(file *os.File) (err error) {
	err = file.Close()
	return err
}

// GetFileLength : How many lines does a text file contain
func GetFileLength(file string) (int, error) {
	i := 0

	lines, err := FileToLines(file)
	if err != nil {
		PrintError("Can't open file: %s", err.Error())
	}
	for range lines {
		i++
	}

	return i, err
}

// FileToLines : Read lines from a text file
func FileToLines(filepath string) ([]string, error) {
	f, err := os.Open(filepath)
	if err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				PrintError(err.Error())
			}
		}()

		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
		return lines, nil
	}
	return nil, err
}

// ExecCmd : exec shell command and put combined output to stdout (line by line)
func ExecCmd(prog string, args string) (int, error) {

	cmd := exec.Command(prog, strings.Split(args, " ")...)

	err := cmd.Start()
	if err != nil {
		return 0, err
	}

	err = cmd.Wait()
	if err != nil {
		PrintError(err.Error())
	}
	return cmd.Process.Pid, err
}

func searchHost(filter string, banner string) bool {
	if strings.Contains(banner, filter) ||
		filter == "" {
		return true
	}

	return false
}

// PrintCyan : print main msg
func PrintCyan(format string, a ...interface{}) {
	color.Set(color.FgCyan)
	defer color.Unset()
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

// PrintRed : print main msg
func PrintRed(format string, a ...interface{}) {
	color.Set(color.FgRed)
	defer color.Unset()
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

// PrintError : print text in red
func PrintError(format string, a ...interface{}) {
	color.Set(color.FgRed, color.Bold)
	defer color.Unset()
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

// PrintSuccess : print text in red
func PrintSuccess(format string, a ...interface{}) {
	color.Set(color.FgHiGreen, color.Bold)
	defer color.Unset()
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

// LogError : print log in red
func LogError(format string, a ...interface{}) {
	color.Set(color.FgRed, color.Bold)
	defer color.Unset()
	log.Printf(format, a...)
	fmt.Print("\n")
}

// LogSuccess : print log in red
func LogSuccess(format string, a ...interface{}) {
	color.Set(color.FgHiGreen, color.Bold)
	defer color.Unset()
	log.Printf(format, a...)
	fmt.Print("\n")
}

// SetCyan : make text following go cyan
func SetCyan() {
	color.Set(color.FgCyan, color.Bold)
}

// UnsetCyan : make text following go back to normal color
func UnsetCyan() {
	color.Unset()
}
