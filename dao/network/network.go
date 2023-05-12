package network

import (
	"fmt"
	"go-web-app/pkg/codeconversion"
	"os/exec"
	"strconv"
	"strings"
)

func NetworkSentSpeed(p string) (s string, err error) {
	switch {
	case p == "uns":
		//网络上传速率
		command := "typeperf -si 1 -sc 1 \"\\Network Interface(*)\\Bytes Sent/sec\"  |findstr \",\" | findstr /V \"Interface\""
		commaninput := exec.Command("powershell.exe", command)
		output, _ := commaninput.CombinedOutput()
		outstring := codeconversion.ConvertByte2String(output, "GB18030")
		lastoutstring := strings.Split(outstring, "\",\"")
		lastoupspeed, _ := strconv.ParseFloat(lastoutstring[1], 8)
		lastoupspeedend := lastoupspeed / 1000
		s = fmt.Sprintf("%.2f", lastoupspeedend)
		return s, err
	case p == "dns":
		//网络下载速率
		command := "typeperf -si 1 -sc 1 \"\\Network Interface(*)\\Bytes Received/sec\"  |findstr \",\" | findstr /V \"Interface\""
		commaninput := exec.Command("powershell.exe", command)
		output, _ := commaninput.CombinedOutput()
		outstring := codeconversion.ConvertByte2String(output, "GB18030")
		lastoutstring := strings.Split(outstring, "\",\"")
		lastoupspeed, _ := strconv.ParseFloat(lastoutstring[1], 8)
		lastoupspeedend := lastoupspeed / 1000
		s = fmt.Sprintf("%.2f", lastoupspeedend)
		return s, err

	}
	return s, err

}
