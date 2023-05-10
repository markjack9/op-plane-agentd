package dao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
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

	ClientParame := models.ClientData{
		ParameterType: parame.ParameterType,
		ClientParame: models.ClientParame{
			Hostid:       parame.Hostid,
			Hostname:     parame.Hostname,
			OptionTime:   parame.OptionTime,
			OptionNote:   parame.OptionNote,
			OptionIp:     parame.OptionIp,
			OpitonParame: parame.OpitonParame,
		},
	}

	clientdata, err := json.Marshal(ClientParame)
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
	fmt.Println(data)
	return
}

func Processing(Hostid int64, cfg *settings.ServerConfig) (err error) {
	fmt.Println(Hostid)
	c := cron.New()
	spec := "*/5 * * * * ?"
	err = c.AddFunc(spec, func() {
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
				OptionIp:     "127.0.0.1",
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
	c.Start()
	select {}
}
