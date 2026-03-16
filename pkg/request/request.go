package request

import (
	"GoSight/config"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type target struct {
	Target  []string          `json:"target"`  //目标地址
	Success []string          `json:"success"` //成功列表
	Config  *config.WebConfig `json:"config"`  //配置项
}

func (this *target) UseConfig(conf string) error { //获取配置项
	if conf == "" {
		return errors.New("配置名不能为空")
	}
	webConfig, ok := config.WebConfigList[conf]
	if !ok {
		return errors.New("配置不存在")
	}
	this.Config = webConfig
	return nil
}

func (this *target) setting(target []string, conf string) error { //设置目标地址和配置
	this.Target = target
	err, v := config.GetWebConf(conf)
	if err != nil {
		if this.Config == nil { //如果获取失败且自己没有配置,则设置默认配置
			this.Config = v
		}
		return err
	}
	this.Config = v
	return nil
}

func (this *target) Start() { //启动任务.理论上应该都使用get方法吧?
	if this.Config == nil {
		this.Config = &config.DefaultWebConfigs
	}

}

// RequestWebPage 根据 WebConfig 配置请求网页，并返回响应体字节或保存的文件路径
func RequestWebPage(urlStr string, webConf *config.WebConfig) ([]byte, string, error) {
	// 1. 创建 HTTP 客户端
	client := &http.Client{
		// 默认处理重定向
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 { // 限制重定向次数，防止无限重定向
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	// 设置超时
	if webConf.Config.Timeout > 0 {
		client.Timeout = time.Duration(webConf.Config.Timeout) * time.Second
	}

	// 处理代理
	if webConf.Agents.Proxy != "" {
		proxyURL, err := url.Parse(webConf.Agents.Proxy)
		if err != nil {
			return nil, "", fmt.Errorf("解析代理地址失败: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	// 2. 创建 HTTP 请求
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加请求头
	if webConf.Agents.Headers != nil {
		for key, values := range webConf.Agents.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	// 添加 Cookie
	if webConf.Agents.Cookies != nil {
		for _, cookie := range webConf.Agents.Cookies {
			req.AddCookie(&cookie)
		}
	}

	// 3. 执行请求并处理重试
	var resp *http.Response
	maxRetries := webConf.Config.MaxRetries
	if maxRetries < 0 { // 确保重试次数不为负
		maxRetries = 0
	}

	for i := 0; i <= maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			break // 请求成功，跳出重试循环
		}

		if i < maxRetries {
			retryDelay := webConf.Config.RetryDelay
			if retryDelay == 0 { // 如果重试延迟为0，则使用超时时间
				retryDelay = webConf.Config.Timeout
			}
			if retryDelay > 0 {
				time.Sleep(time.Duration(retryDelay) * time.Millisecond)
			}
		} else {
			if err != nil {
				return nil, "", fmt.Errorf("请求失败，已达最大重试次数: %w", err)
			}
			return nil, "", fmt.Errorf("请求返回非成功状态码 %d，已达最大重试次数", resp.StatusCode)
		}
	}
	defer resp.Body.Close()

	// 4. 处理响应
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	// 判断是否为文件类型（图片、压缩包等）
	if isFileContentType(contentType) {
		// 获取文件名
		fileName := getFileNameFromURL(urlStr)
		if fileName == "" {
			fileName = "downloaded_file" // 默认文件名
		}

		// 确保输出目录存在
		outputDir := webConf.Config.Output
		if outputDir == "" {
			outputDir = "./downloads" // 默认下载目录
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, "", fmt.Errorf("创建输出目录失败: %w", err)
		}

		filePath := path.Join(outputDir, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			return nil, "", fmt.Errorf("创建文件失败: %w", err)
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("保存文件失败: %w", err)
		}
		return nil, filePath, nil // 返回文件路径
	} else {
		// 返回响应体字节
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("读取响应体失败: %w", err)
		}
		return bodyBytes, "", nil
	}
}

// isFileContentType 判断 Content-Type 是否为文件类型
func isFileContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	// 常见的图片、压缩包、二进制文件类型
	if strings.HasPrefix(contentType, "image/") ||
		strings.HasPrefix(contentType, "application/zip") ||
		strings.HasPrefix(contentType, "application/x-rar-compressed") ||
		strings.HasPrefix(contentType, "application/pdf") ||
		strings.HasPrefix(contentType, "application/octet-stream") ||
		strings.HasPrefix(contentType, "application/vnd.openxmlformats-officedocument") || // Office 文件
		strings.HasPrefix(contentType, "audio/") ||
		strings.HasPrefix(contentType, "video/") {
		return true
	}
	return false
}

// getFileNameFromURL 从 URL 中提取文件名
func getFileNameFromURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return path.Base(parsedURL.Path)
}
