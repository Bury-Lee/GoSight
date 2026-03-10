package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// 定义请求/响应的数据结构
type RequestData struct {
	Message string `json:"message"`
}

type ResponseData struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

func main() {
	// 设置路由
	http.HandleFunc("/", indexHandler)            // 首页
	http.HandleFunc("/api/data", apiHandler)      // API 接口
	http.HandleFunc("/api/submit", submitHandler) // 接收前端数据

	// 启动服务器
	port := ":8080"
	fmt.Printf("服务器启动在 http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// 首页处理器 - 返回 HTML 页面
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Go Web 应用</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        button { padding: 10px 20px; margin: 10px; cursor: pointer; }
        #result { margin-top: 20px; padding: 15px; background: #f0f0f0; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>Go 语言 Web 应用示例</h1>
    
    <h2>1. 获取后端数据</h2>
    <button onclick="fetchData()">获取数据</button>
    
    <h2>2. 发送数据到后端</h2>
    <input type="text" id="inputMsg" placeholder="输入消息" style="padding: 8px; width: 200px;">
    <button onclick="sendData()">提交</button>
    
    <div id="result">结果将显示在这里...</div>

    <script>
        // 从后端获取数据
        async function fetchData() {
            try {
                const response = await fetch('/api/data');
                const data = await response.json();
                document.getElementById('result').innerHTML = 
                    '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
            } catch (error) {
                document.getElementById('result').innerHTML = '错误: ' + error;
            }
        }

        // 发送数据到后端
        async function sendData() {
            const msg = document.getElementById('inputMsg').value;
            try {
                const response = await fetch('/api/submit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ message: msg })
                });
                const data = await response.json();
                document.getElementById('result').innerHTML = 
                    '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
            } catch (error) {
                document.getElementById('result').innerHTML = '错误: ' + error;
            }
        }
    </script>
</body>
</html>
	`
	fmt.Fprint(w, html)
}

// API 处理器 - 返回 JSON 数据
func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := ResponseData{
		Status:  "success",
		Message: "Hello from Go backend!",
		Time:    fmt.Sprintf("%d", 1234567890),
	}

	json.NewEncoder(w).Encode(response)
}

// 接收前端提交的 JSON 数据
func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	var req RequestData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "解析 JSON 失败", http.StatusBadRequest)
		return
	}

	// 处理接收到的数据
	response := ResponseData{
		Status:  "received",
		Message: "收到消息: " + req.Message,
		Time:    "now",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
