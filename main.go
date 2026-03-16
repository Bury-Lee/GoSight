package main

import (
	"GoSight/config"
	"GoSight/globel"
	"GoSight/logs"
	"GoSight/req_res"
	"fmt"
	"net/http"
)

func main() {
	err := config.Config.Init(config.ConfigPath)
	if err != nil {
		fmt.Println("初始化配置文件失败:", err)
		return
	}

	logs.InitLog()

	defer config.DefaultWebConfigs.SaveAsJson()
	defer globel.LogFile.Close() // 确保主程序退出时关闭日志文件

	//debug
	myConfig := &config.WebConfig{
		RootName:   "root_job",
		ConfigName: "spider_v1",
		Config: config.BaseConfig{
			Output:      "/data/output.json",
			Concurrency: 10,
			Delay:       1000, // 1秒
			Timeout:     30,   // 30秒
			MaxRetries:  3,
		},
		Web: config.BlackConfig{
			BlackList: []string{"192.168.1.1", "bad-site.com"},
		},
		Agents: config.Agent{
			Headers: http.Header{
				"User-Agent": []string{"MyCustomSpider/1.0"},
			},
		},
		Render: config.RenderConfig{
			Enable:   true,
			Engine:   "chromium",
			Headless: true,
		},
		Next:     []*config.WebConfig{},
		NextName: []string{},
	}
	taget := req_res.Target{
		Target: []string{"http://127.0.0.1:8081/web/login.html"},
		Body:   "懂啊打动",
		Config: myConfig,
	}
	result1, result2, err := taget.RequestWebPage(taget.Target[0], myConfig)
	fmt.Print(string(result1), result2)
	//debug
}
