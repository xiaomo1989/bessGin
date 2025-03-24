package main

import (
	"bessGin/config"
	"bessGin/route"
	"encoding/json"
	"fmt"
)

func main() {
	//加载配置
	var config config.Config
	conf := config.GetConf()

	//将对象，转换成json格式
	data_config, err := json.Marshal(conf)

	if err != nil {
		fmt.Println("err:\t", err.Error())
		return
	}
	fmt.Println("data_config:\t", string(data_config))
	//fmt.Println("config.Database.Driver=",config.Database.Driver);
	//加载路由
	r := route.InitRouter()
	r.Run(":8817")

}
