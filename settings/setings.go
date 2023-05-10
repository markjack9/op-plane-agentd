package settings

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Conf = new(AppConfig)

type AppConfig struct {
	Name          string `mapstructure:"name"`
	Mode          string `mapstructure:"mode"`
	Version       string `mapstructure:"version"`
	Port          int    `mapstructure:"port"`
	StartTime     string `mapstructure:"start_time"`
	MachineId     int64  `mapstructure:"machine_id"`
	*LogConfig    `mapstructure:"log"`
	*ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Ip        string `mapstructure:"ip"`
	Port      int    `mapstructure:"port"`
	Heartbeat int    `mapstructure:"heartbeat"`
	HostName  string `mapstructure:"hostname"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

func Init(configfile string) (err error) {
	viper.SetConfigFile(configfile)
	//指定配置文件
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Profile read failed, please specify the configuration file:%v\n", err)
		return
	}
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改...")
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
		}
	})
	return
}
