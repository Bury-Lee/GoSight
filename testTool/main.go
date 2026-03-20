package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RequestLogMiddleware 用于打印完整的请求信息（Header, Cookie, Body）
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 打印请求方法和 URL
		fmt.Println("========== 收到新请求 ==========")
		fmt.Printf("Method: %s\n", c.Request.Method)
		fmt.Printf("URL: %s\n", c.Request.URL.String())
		fmt.Printf("Remote Addr: %s\n", c.ClientIP())

		// 2. 打印所有请求头 (Headers)，Cookie 也包含在其中
		fmt.Println("--- Request Headers ---")
		for key, values := range c.Request.Header {
			// 将值数组转换为字符串打印
			fmt.Printf("%s: %s\n", key, strings.Join(values, ", "))
		}

		// // 单独强调一下 Cookie (可选，因为上面已经打印了所有 Header)
		// if cookies := c.Request.Cookies(); len(cookies) > 0 {
		// 	fmt.Println("--- 解析后的 Cookies ---")
		// 	for _, cookie := range cookies {
		// 		fmt.Printf("Cookie Name: %s, Value: %s\n", cookie.Name, cookie.Value)
		// 	}
		// }

		// 3. 读取并打印请求体 (Body)
		// 注意：Body 是一个 io.ReadCloser，读取后流会耗尽，必须重新赋值以便后续 handler 使用
		var bodyBytes []byte
		var err error

		if c.Request.Body != nil {
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				fmt.Printf("读取请求体失败: %v\n", err)
				c.Next()
				return
			}

			// 打印请求体内容
			fmt.Println("--- Request Body ---")
			// 尝试判断是否为二进制或文本，这里简单打印字符串，如果是图片/文件可能会乱码
			// 对于 static 文件请求，通常没有 Body 或者 Body 为空
			if len(bodyBytes) > 0 {
				// 简单处理：如果内容太长，只打印前 2000 个字符避免刷屏
				bodyStr := string(bodyBytes)
				if len(bodyStr) > 2000 {
					fmt.Println(bodyStr[:2000])
					fmt.Println("... (内容过长，已截断)")
				} else {
					fmt.Println(bodyStr)
				}
			} else {
				fmt.Println("(空)")
			}

			// 【关键步骤】将读取过的数据重新写回 Body，否则后续的 gin handler 无法读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		} else {
			fmt.Println("--- Request Body ---")
			fmt.Println("(无请求体)")
		}

		fmt.Println("============================")

		// 继续处理后续的逻辑
		c.Next()
	}
}

func main() {
	r := gin.Default()

	// 配置 CORS (保持原有逻辑)
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 开发环境允许所有来源
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Cookie"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// 【新增】注册日志中间件，放在最前面以确保捕获所有请求
	r.Use(RequestLogMiddleware())

	// 静态文件服务
	// 当访问 http://localhost:8081/web/xxx 时，Gin 会去 "./static" 目录下寻找 xxx 文件
	r.Static("", "./static")

	// 添加一个测试 POST 接口，方便你测试带 Body 的请求
	r.POST("/test-api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "请求已接收，请查看控制台输出的详细日志",
			"path":    c.Request.URL.Path,
		})
	})

	fmt.Println("服务器启动于: http://127.0.0.1:8081")
	// 启动服务器
	if err := r.Run("127.0.0.1:8081"); err != nil {
		panic(err)
	}
}
