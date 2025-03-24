package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	Database Database `yaml:"database"`
}

type Database struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

// 读取Yaml配置文件，并转换成Config对象  struct结构
func (config *Config) GetConf() *Config {
	//获取项目的执行路径
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println("path=", path)
	vip := viper.New()
	vip.AddConfigPath(path + "/config") //设置读取的文件路径
	vip.SetConfigName("settings")       //设置读取的文件名
	vip.SetConfigType("yaml")           //设置文件的类型
	//尝试进行配置读取
	if err := vip.ReadInConfig(); err != nil {
		panic(err)
	}

	err = vip.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	return config
}
