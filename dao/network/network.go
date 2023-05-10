package network

import (
	"fmt"
	"go-web-app/models"
	"go-web-app/pkg/codeconversion"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

func NetworkSentSpeed(p *models.ParamSystemGet) (s string, err error) {
	switch {
	case p.ParameterType == "uns":
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
	case p.ParameterType == "dns":
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
	case p.ParameterType == "mp":
		//内存使用率
		command := "typeperf -si 1 -sc 1 \"\\Memory\\% Committed Bytes In Use\" |findstr /V \"Memory\" |findstr /V \"?\""
		commaninput := exec.Command("powershell.exe", command)
		output, _ := commaninput.CombinedOutput()
		outstring := codeconversion.ConvertByte2String(output, "GB18030")
		lastoutstring := strings.Split(outstring, ",\"")
		lastoutstringdel := strings.Split(lastoutstring[1], "\"")
		lastoupspeed, err := strconv.ParseFloat(lastoutstringdel[0], 8)
		s = fmt.Sprintf("%.0f", math.Floor(lastoupspeed))
		return s, err
	}
	return s, err

}
