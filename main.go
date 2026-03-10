package main

import (
	"GoSight/config"
	"GoSight/logService"
	"fmt"
)

func main() {
	err := config.Config.Init(config.ConfigPath)
	if err != nil {
		fmt.Println("初始化配置文件失败:", err)
		return
	}

	logService.InitLog()

	defer config.DefaultWebConfigs.SaveAsGob()
	defer logService.LogFile.Close() // 确保主程序退出时关闭日志文件
}
