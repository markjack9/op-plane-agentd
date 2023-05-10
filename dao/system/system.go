package system

import (
	"fmt"
	"go-web-app/pkg/codeconversion"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

func Systeminfo(p string) (s string, err error) {
	switch {
	case p == "cpu":
		//cpu使用率
		command := " typeperf -si 1 -sc 1 \"\\Processor(_Total)\\% Processor Time\" |findstr /V \"Processor\" |findstr /V \"?\" "
		commaninput := exec.Command("powershell.exe", command)
		output, _ := commaninput.CombinedOutput()
		outstring := codeconversion.ConvertByte2String(output, "GB18030")
		lastoutstring := strings.Split(outstring, ",\"")
		lastoutstringdel := strings.Split(lastoutstring[1], "\"")
		lastoupspeed, _ := strconv.ParseFloat(lastoutstringdel[0], 8)
		s = fmt.Sprintf("%.0f", math.Floor(lastoupspeed))
		return s, err
	case p == "mp":
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
	case p == "dt":
		//系统磁盘容量
		command := "wmic LogicalDisk where \"Caption='C:'\" get  Size /value | findstr \"Size\""
		commaninput := exec.Command("powershell.exe", command)
		output, _ := commaninput.CombinedOutput()
		outstring := codeconversion.ConvertByte2String(output, "GB18030")
		lastoutstring := strings.Split(outstring, "=")
		b := strings.Replace(lastoutstring[1], "\r\n", "", -1)
		c, err := strconv.ParseInt(b, 10, 64)
		d := fmt.Sprintf("%.0f", math.Floor(float64(c/1073741824)))
		return d, err
	case p == "fdp":
		//系统磁盘剩余空间占总比的
		command := "typeperf -si 1 -sc 1 \"\\LogicalDisk(C:)\\% Free Space\" |findstr /V \"Space\"| findstr /V \"?\""
		commaninput := exec.Command("powershell.exe", command)
		output, _ := commaninput.CombinedOutput()
		outstring := codeconversion.ConvertByte2String(output, "GB18030")
		lastoutstring := strings.Split(outstring, ",\"")
		lastoutstringdel := strings.Split(lastoutstring[1], "\"")
		lastoupspeed, _ := strconv.ParseFloat(lastoutstringdel[0], 8)
		s = fmt.Sprintf("%.0f", math.Floor(lastoupspeed))
		return s, err
	case p == "mt":
		//系统总内存
		command := " wmic ComputerSystem get TotalPhysicalMemory | findstr /V \"Total\""
		commaninput := exec.Command("powershell.exe", command)
		output, _ := commaninput.CombinedOutput()
		outstring := codeconversion.ConvertByte2String(output, "GB18030")
		a := strings.Replace(outstring, "\r\n", "", -1)
		a = strings.Replace(a, " ", "", -1)
		b, _ := strconv.ParseInt(a, 10, 64)
		s := ((b / 1073741824) + 1)
		d := strconv.FormatInt(s, 10)
		return d, err
	case p == "sup":
		//系统运行时间
		Time := []string{"Days", "Hours", "Minutes", "Seconds"}
		var TimeString []string
		for _, v := range Time {
			command := "(get-date) - (gcim Win32_OperatingSystem).LastBootUpTime | findstr /V  Total |findstr " + v
			commaninput := exec.Command("powershell.exe", command)
			output, _ := commaninput.CombinedOutput()
			outstring := codeconversion.ConvertByte2String(output, "GB18030")
			lastoutstring := strings.Split(outstring, ":")
			a := strings.Replace(lastoutstring[1], "\r\n", "", -1)
			TimeString = append(TimeString, a+v)
		}
		s := strings.Join(TimeString, "")
		s = strings.Replace(s, " ", "", -1)
		s = strings.Replace(s, "Days", "天", -1)
		s = strings.Replace(s, "Hours", "小时", -1)
		s = strings.Replace(s, "Minutes", "分钟", -1)
		s = strings.Replace(s, "Seconds", "秒", -1)
		return s, err
	}
	return s, err

}
