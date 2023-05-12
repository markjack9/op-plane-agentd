package dao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"go-web-app/dao/network"
	"go-web-app/dao/system"
	"go-web-app/models"
	"go-web-app/pkg/todaytime"
	"go-web-app/settings"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func ServerConfirm(cfg *settings.ServerConfig) (Hostid int64, err error) {
	url := fmt.Sprintf("http://%s:%d/clientdata", cfg.Ip, cfg.Port)

	ClientParame := models.ClientData{
		ParameterType: "Confirm",
		ClientParame: models.ClientParame{
			Hostid:   0,
			Hostname: cfg.HostName,
		},
	}
	clientdata, err := json.Marshal(ClientParame)
	if err != nil {
		return
	}
	reader := bytes.NewReader(clientdata)
	// 1. 创建http客户端实例
	client := &http.Client{}
	// 2. 创建请求实例
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		panic(err)
	}
	// 3. 设置请求头，可以设置多个
	req.Header.Set("Content-Type", "application/json")

	// 4. 发送请求
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	var data models.ServerResult
	resutbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(strings.NewReader(string(resutbody)))
	decoder.UseNumber()
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("error:", err)
	}

	s := fmt.Sprintf("%v", data.Data)
	Hostid, _ = strconv.ParseInt(s, 10, 64)
	fmt.Println("获取机器Id", Hostid)
	return
}

func GoPost(parame models.ClientData, cfg *settings.ServerConfig) (err error) {
	url := fmt.Sprintf("http://%s:%d/clientdata", cfg.Ip, cfg.Port)
	clientdata, err := json.Marshal(parame)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(clientdata)

	// 1. 创建http客户端实例
	client := &http.Client{}
	// 2. 创建请求实例
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		panic(err)
	}
	// 3. 设置请求头，可以设置多个
	req.Header.Set("Content-Type", "application/json")
	fmt.Println("已发送请求")
	// 4. 发送请求
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	var data models.ServerResult
	resutbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(strings.NewReader(string(resutbody)))
	decoder.UseNumber()
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("error:", err)
	}
	return
}

func Processing(Hostid int64, cfg *settings.AppConfig) (err error) {
	systeminfo, err := system.Systeminfo("systeminfo")
	fmt.Println(systeminfo)
	parame := models.ClientData{
		ParameterType: "systeminfo",
		ClientParame: models.ClientParame{
			Hostid:       Hostid,
			Hostname:     cfg.HostName,
			OptionTime:   todaytime.NowTimeFull(),
			OptionNote:   "",
			OptionIp:     cfg.ClientIp,
			OpitonParame: systeminfo,
		},
	}

	err = GoPost(parame, settings.Conf.ServerConfig)
	if err != nil {
		return
	}
	c := cron.New()
	err = c.AddFunc("* */10 * * *", func() {
		systemdata, err := system.Systeminfo("sup")
		if err != nil {
			return
		}
		parame := models.ClientData{
			ParameterType: "uptime",
			ClientParame: models.ClientParame{
				Hostid:       Hostid,
				Hostname:     cfg.HostName,
				OptionTime:   todaytime.NowTimeFull(),
				OptionNote:   "",
				OptionIp:     cfg.ClientIp,
				OpitonParame: systemdata,
			},
		}

		err = GoPost(parame, settings.Conf.ServerConfig)
		if err != nil {
			return
		}
	})
	if err != nil {
		return err
	}
	err = c.AddFunc("*/1 * * * *", func() {
		cpu, err := system.Systeminfo("cpu")
		if err != nil {
			return
		}
		memory, err := system.Systeminfo("mp")
		if err != nil {
			return
		}
		uns, err := network.NetworkSentSpeed("uns")
		if err != nil {
			return
		}
		dns, err := network.NetworkSentSpeed("dns")
		if err != nil {
			return
		}

		parame := models.ClientData{
			ParameterType: "basemonitoring",
			ClientParame: models.ClientParame{
				Hostid:             Hostid,
				Hostname:           cfg.HostName,
				OptionTime:         todaytime.NowTimeFull(),
				OptionIp:           cfg.ClientIp,
				OptionParameCpu:    cpu,
				OptionParameMemory: memory,
				OptionParameUns:    uns,
				OptionParameDns:    dns,
			},
		}
		fmt.Println("定时任务", parame)
		err = GoPost(parame, settings.Conf.ServerConfig)
		if err != nil {
			return
		}
	})
	if err != nil {
		return err
	}
	c.Start()
	select {}
}
