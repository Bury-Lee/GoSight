package main

import (
	"GoSight/config"
	global "GoSight/globel"
	"GoSight/logs"
	"GoSight/req_res"
	"fmt"
)

func main() {
	err := config.Config.Init(config.ConfigPath)
	if err != nil {
		fmt.Println("初始化配置文件失败:", err)
		return
	}

	logs.InitLog()

	defer config.DefaultWebConfig.SaveAsJson()
	defer global.LogFile.Close() // 确保主程序退出时关闭日志文件

	//debug

	taget := req_res.Target{
		Target: []string{"http://127.0.0.1:8081/web/login.html"},
		Body:   "火影,启动!",
		Config: &config.DefaultWebConfig,
	}
	taget.Start()
	// config.AllLoad("D:/并行爬虫库/UserData")
	//debug

	//debug
	// ... 在 main 函数中
	// _, err = config.LoadFromJSON("UserData/default_spider copy.json")
	// if err != nil {
	// 	print(err)
	// }

	// 将结构体转换为漂亮的 JSON 字符串
	// jsonData, err := json.MarshalIndent(config.DefaultWebConfig, "", "  ")
	// if err != nil {
	// 	fmt.Println("JSON 转换错误:", err)
	// } else {
	// 	fmt.Println(string(jsonData))
	// }
	//debug
}
