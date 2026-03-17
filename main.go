package main

import (
	"GoSight/config"
	global "GoSight/globel"
	"GoSight/logs"
	"encoding/json"
	"fmt"
)

func main() {
	err := config.Config.Init(config.ConfigPath)
	if err != nil {
		fmt.Println("初始化配置文件失败:", err)
		return
	}

	logs.InitLog()

	defer config.DefaultWebConfigs.SaveAsJson()
	defer global.LogFile.Close() // 确保主程序退出时关闭日志文件

	//debug
	// myConfig := &config.WebConfig{
	// 	RootName:   "root_job",
	// 	ConfigName: "spider_v1",
	// 	Config: config.BaseConfig{
	// 		Output:      "/data/output.json",
	// 		Concurrency: 10,
	// 		Delay:       1000, // 1秒
	// 		Timeout:     30,   // 30秒
	// 		MaxRetries:  3,
	// 	},
	// 	Web: config.BlackConfig{
	// 		BlackList: []string{"192.168.1.1", "bad-site.com"},
	// 	},
	// 	Agents: config.Agent{
	// 		Headers: http.Header{
	// 			"User-Agent": []string{"原神/1.0"},
	// 		},
	// 	},
	// 	Render: config.RenderConfig{
	// 		Enable:   true,
	// 		Engine:   "原神omium",
	// 		Headless: true,
	// 	},
	// 	// 在这里嵌套子层
	// 	Next: []*config.WebConfig{
	// 		// --- 子层 1: 列表页抓取 (List Page) ---
	// 		// 特点：不需要渲染引擎（假设列表是静态的），提高并发，缩短超时
	// 		{
	// 			RootName:   "list_layer",
	// 			ConfigName: "spider_v1_list",
	// 			Config: config.BaseConfig{
	// 				Output:      "/data/list_output.json",
	// 				Concurrency: 20,  // 列表页通常可以更高并发
	// 				Delay:       500, // 0.5秒
	// 				Timeout:     15,  // 15秒足够
	// 				MaxRetries:  2,
	// 			},
	// 			Web: config.BlackConfig{
	// 				// 继承或覆盖黑名单，这里添加一个特定的测试IP
	// 				BlackList: []string{"10.0.0.1"},
	// 			},
	// 			Agents: config.Agent{
	// 				Headers: http.Header{
	// 					"User-Agent": []string{"Mozilla/5.0 (ListBot)"},
	// 				},
	// 			},
	// 			Render: config.RenderConfig{
	// 				Enable:   false, // 列表页通常不需要JS渲染，关闭以节省资源
	// 				Engine:   "",
	// 				Headless: false,
	// 			},
	// 			// 可以继续嵌套下一级
	// 			Next:     []*config.WebConfig{},
	// 			NextName: []string{},
	// 		},

	// 		// --- 子层 2: 详情页抓取 (Detail Page) ---
	// 		// 特点：需要渲染引擎（动态内容），较低的并发以防被封，特定的Header
	// 		{
	// 			RootName:   "detail_layer",
	// 			ConfigName: "spider_v1_detail",
	// 			Config: config.BaseConfig{
	// 				Output:      "/data/detail_output.json",
	// 				Concurrency: 5,    // 详情页负载高，降低并发
	// 				Delay:       2000, // 2秒，模拟人类行为
	// 				Timeout:     60,   // 60秒，等待JS加载
	// 				MaxRetries:  5,    // 重要数据，多重试几次
	// 			},
	// 			Web: config.BlackConfig{
	// 				BlackList: []string{}, // 清空黑名单或使用默认的
	// 			},
	// 			Agents: config.Agent{
	// 				Headers: http.Header{
	// 					"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; DetailBot)"},
	// 					"Accept-Language": []string{"zh-CN,zh;q=0.9"},
	// 				},
	// 			},
	// 			Render: config.RenderConfig{
	// 				Enable:   true, // 必须开启渲染
	// 				Engine:   "原神omium",
	// 				Headless: true,
	// 			},
	// 			// 指向下一层：数据接口
	// 			Next:     []*config.WebConfig{},
	// 			NextName: []string{},
	// 		},

	// 		// --- 子层 3: 纯数据接口/API (API Layer) ---
	// 		// 特点：极速，无渲染，高并发，直接抓取JSON
	// 		{
	// 			RootName:   "api_layer",
	// 			ConfigName: "spider_v1_api",
	// 			Config: config.BaseConfig{
	// 				Output:      "/data/final_json.json",
	// 				Concurrency: 50,  // API通常很快，可以很高并发
	// 				Delay:       100, // 100毫秒
	// 				Timeout:     10,  // 10秒超时
	// 				MaxRetries:  1,   // 失败即放弃，避免阻塞
	// 			},
	// 			Web: config.BlackConfig{
	// 				BlackList: []string{"slow-api.com"},
	// 			},
	// 			Agents: config.Agent{
	// 				Headers: http.Header{
	// 					"User-Agent":   []string{"Go-HttpClient/2.0"},
	// 					"Content-Type": []string{"application/json"},
	// 				},
	// 			},
	// 			Render: config.RenderConfig{
	// 				Enable:   false, // API不需要渲染
	// 				Engine:   "",
	// 				Headless: false,
	// 			},
	// 			// 最后一层，不再嵌套
	// 			Next:     nil,
	// 			NextName: nil,
	// 		},
	// 	},
	// 	// 如果需要给这一层的 next 命名，可以在这里填写
	// 	NextName: []string{"to_list", "to_detail", "to_api"},
	// }

	// taget := req_res.Target{
	// 	Target: []string{"http://127.0.0.1:8081/web/login.html"},
	// 	Body:   "火影,启动!",
	// 	Config: myConfig,
	// }
	// result1, result2, err := taget.RequestWebPage(taget.Target[0], myConfig)
	// fmt.Print(string(result1), result2)
	// config.AllLoad("D:/并行爬虫库/UserData")
	// myConfig.Add()
	// myConfig.SaveAsJson()
	//debug

	//debug
	// ... 在 main 函数中
	_, err = config.LoadFromJSON("UserData/default_spider copy.json")
	if err != nil {
		print(err)
	}

	// 将结构体转换为漂亮的 JSON 字符串
	jsonData, err := json.MarshalIndent(config.DefaultWebConfigs, "", "  ")
	if err != nil {
		fmt.Println("JSON 转换错误:", err)
	} else {
		fmt.Println(string(jsonData))
	}
	//debug
}
