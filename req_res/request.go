// 处理构建请求体和响应体
package req_res

import (
	"GoSight/config"
	global "GoSight/globel"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
)

func (this *Target) UseConfig(conf string) error { //获取配置项
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

func (this *Target) setting(target []string, conf string) error { //设置目标地址和配置
	//如果当前配置没有就递归的向上一级寻找配置
	this.Target = target
	err, v := config.GetWebConf(conf)
	if err != nil || v == nil {
		if this.Config == nil { //如果获取失败且自己没有配置,则设置默认配置
			this.Config = v
		}
		return err
	}
	this.Config = v
	return nil
}

// Target 结构体定义 (基于你的描述补充完整)
type Target struct {
	Target  []string          `json:"target"`
	Body    any               `json:"body"`
	Success []string          `json:"success"`
	Config  *config.WebConfig `json:"config"`
}

type Tasks struct { //任务组
	Target []Target `json:"target"`
	Name   []string `json:"name"` //任务名
}

func (this *Target) SmartStart() { //智能保存文件并将文件本地路径填入到文件地址中

}

// Start 启动爬取任务
func (this *Target) Start() {
	// 1. 初始化配置
	if this.Config == nil {
		// 假设 DefaultWebConfig 已在全局定义，如果没有需自行定义一个空的
		this.Config = &config.DefaultWebConfig
	}

	// 深拷贝或引用当前配置 (根据你的继承逻辑，这里假设 conf 已经是合并好后的最终配置)
	// 注意：如果 this.Config 是指针且后续会被修改，这里可能需要 lock 或者 deepcopy
	conf := *this.Config
	point := conf.Root
	// 假设 point 是一个包含配置树的对象，且 point.Root 是根节点配置
	// 如果 point 未初始化，你需要先确保它指向正确的配置上下文
	if point == nil {
		// 处理 point 为空的情况，或者直接报错/返回
		point = &config.DefaultWebConfig
	}
	// 1. 检查整个 WebConfig 是否为空
	if reflect.DeepEqual(conf, config.EmptyWebConfig) { //当conf为默认配置时不可能为空
		global.Logger.Warn("配置为空, 自动使用默认配置")
		conf = config.DefaultWebConfig
	} else { //当conf为默认配置时以下配置项也不可能为空
		point = conf.Root
		for reflect.DeepEqual(conf.Agents, config.EmptyAgent) {
			global.Logger.Debug("Agents 配置为空, 使用根节点配置")
			if point == nil {
				global.Logger.Error("配置上下文 point 为空")
				point = &config.DefaultWebConfig
			}
			conf.Agents = point.Agents
		}
		point = conf.Root
		for reflect.DeepEqual(conf.RenderConfig, config.EmptyRenderConfig) {
			global.Logger.Debug("RenderConfig 配置为空, 使用根节点配置")
			if point == nil {
				global.Logger.Error("配置上下文 point 为空")
				point = &config.DefaultWebConfig
			}
			conf.RenderConfig = point.RenderConfig
			point = point.Root
		}
		point = conf.Root //复位
		for reflect.DeepEqual(conf.BaseConfig, config.EmptyBaseConfig) {
			global.Logger.Debug("BaseConfig 配置为空, 使用根节点配置")
			if point == nil {
				global.Logger.Error("配置上下文 point 为空")
				point = &config.DefaultWebConfig
			}
			conf.BaseConfig = point.BaseConfig
			point = point.Root
		}
	}

	// 初始化日志 (简单示例，实际应从 config.Log 初始化 slog)
	global.Logger.Info("任务开始")

	if len(this.Target) == 0 {
		global.Logger.Warn("无任务")
		return
	}

}

// RequestWebPage 根据 WebConfig 配置请求网页，并返回响应体字节或保存的文件路径
func (this *Target) RequestWebPage(urlStr string, webConf *config.WebConfig) ([]byte, string, error) { //TODO:自定义正则化保存
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
	if webConf.BaseConfig.Timeout > 0 {
		client.Timeout = time.Duration(webConf.BaseConfig.Timeout) * time.Second
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

	var err error
	var req *http.Request
	// 2. 创建 HTTP 请求
	tem, err := json.Marshal(this.Body)
	if err != nil { //如果序列化失败就是不启用请求体
		global.Logger.Error("获取请求体失败: " + err.Error())
		req, err = http.NewRequest("GET", urlStr, nil)
	} else {
		body := strings.NewReader(string(tem))
		req, err = http.NewRequest("GET", urlStr, body)
	}
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
	maxRetries := webConf.BaseConfig.MaxRetries
	if maxRetries < 0 { // 确保重试次数不为负
		maxRetries = 0
	}

	for i := 0; i <= maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			break // 请求成功，跳出重试循环
		}

		if i < maxRetries {
			retryDelay := webConf.BaseConfig.RetryDelay
			if retryDelay == 0 { // 如果重试延迟为0，则使用超时时间
				retryDelay = webConf.BaseConfig.Timeout
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
			fileName = "downloaded_file" // TODO:改正保存文件名
		}

		// 确保输出目录存在
		outputDir := webConf.BaseConfig.Output
		if outputDir == "" {
			outputDir = config.DefaultWebConfig.BaseConfig.Output // 默认下载目录
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
func downloadFile(url string, filepath string) error { //原始文件下载函数,之后在智能处理网页文件镶嵌时可能会用
	// 1. 发起请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 2. 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// 3. 创建本地文件
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 4. 将响应体写入文件 (流式拷贝)
	_, err = io.Copy(out, resp.Body)
	return err
}
