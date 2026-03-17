// 处理构建请求体和响应体
package req_res

import (
	"GoSight/config"
	global "GoSight/globel"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// type Target struct {
// 	Target []string `json:"target"` //目标地址
// 	Body   any      `json:"body"`   //对这些目标设置同一请求体.如果希望有不同请求体就应该使用不同的target
// 	//理论上不应该有上传文件环节
// 	Success []string          `json:"success"` //成功列表
// 	Config  *config.WebConfig `json:"config"`  //配置项        自己存储一个或使用指针?也许指针的同步性更好
// }

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

// func (this *Target) Start() { //启动任务.理论上应该都使用get方法吧?
// 	if this.Config == nil {
// 		this.Config = &config.DefaultWebConfig
// 	}
// 	var conf config.WebConfig = *this.Config //本次使用的配置
// 	// tem := config.WebConfig{}
// 	// if conf.Agents == tem.Agents {
// 	// 	conf.Agents = conf.Root.Agents
// 	// } //构建栈,进行溯源后增量更新的形式?从上到下进行继承?
// 	for _,
// }

// Target 结构体定义 (基于你的描述补充完整)
type Target struct {
	Target  []string          `json:"target"`
	Body    any               `json:"body"`
	Success []string          `json:"success"`
	Config  *config.WebConfig `json:"config"`
	mu      sync.Mutex        // 用于保护 Success 切片的并发写入
	// logger  *slog.Logger      // 建议注入 logger，这里暂用默认或从 config 获取
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

	// 初始化日志 (简单示例，实际应从 config.Log 初始化 slog)
	global.Logger.Info("任务开始")

	if len(this.Target) == 0 {
		global.Logger.Warn("无任务")
		return
	}

	// 2. 并发控制设置
	concurrency := conf.Config.Concurrency
	if concurrency <= 0 {
		concurrency = 1 // 默认至少为 1
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	global.Logger.Info("Starting crawl task",
		"total_targets", len(this.Target),
		"concurrency", concurrency,
		"max_retries", conf.Config.MaxRetries)

	// 3. 遍历目标 URL
	for _, urlStr := range this.Target {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			this.executeWithRetry(url, &conf)
		}(urlStr)
	}

	wg.Wait()
	global.Logger.Info("任务完成", "完成数量:", len(this.Success))
}

// executeWithRetry 处理单个 URL 的重试、延迟和请求逻辑
func (this *Target) executeWithRetry(urlStr string, conf *config.WebConfig) {
	cfg := conf.Config // 基础配置

	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	var lastErr error
	var body []byte
	var contentType string

	// 重试循环: 尝试次数 = 1 (首次) + maxRetries
	for attempt := 0; attempt <= maxRetries; attempt++ {

		// --- A. 延迟处理 (Delay & RandomDelay) ---
		// 如果不是第一次尝试 (即重试时)，或者配置了初始延迟
		// 通常首次请求也建议加上基础延迟以防瞬间爆发
		sleepDuration := this.calculateDelay(cfg, attempt)

		if sleepDuration > 0 {
			global.Logger.Debug("Sleeping before request", "url", urlStr, "duration_ms", sleepDuration.Milliseconds(), "attempt", attempt+1)
			time.Sleep(sleepDuration)
		}

		// --- B. 执行请求 ---
		global.Logger.Debug("Requesting URL", "url", urlStr, "attempt", attempt+1)

		// 调用你已实现的函数
		body, contentType, lastErr = this.RequestWebPage(urlStr, conf)

		// --- C. 结果判断 ---
		if lastErr == nil {
			// 成功
			global.Logger.Info("Request successful", "url", urlStr, "content_type", contentType, "size_bytes", len(body))

			// 线程安全地添加到成功列表
			this.mu.Lock()
			this.Success = append(this.Success, urlStr)
			this.mu.Unlock()

			// 在这里可以处理 body 数据，例如保存文件 (根据 cfg.Output)
			// saveToFile(urlStr, body, cfg.Output)
			return // 成功则退出重试循环
		}

		// 失败处理
		global.Logger.Warn("Request failed", "url", urlStr, "attempt", attempt+1, "error", lastErr.Error())

		// 如果是最后一次尝试仍然失败，跳出循环
		if attempt == maxRetries {
			break
		}

		// --- D. 重试退避策略 (Backoff) ---
		// 在下次重试前等待。注意：calculateDelay 已经处理了基础 delay，
		// 这里通常专门处理因失败而产生的额外退避等待，或者将两者结合。
		// 为了简化，上面的 calculateDelay 已经包含了基于 attempt 的退避计算。
		// 如果需要区分 "请求间正常延迟" 和 "失败后惩罚延迟"，可在此处额外增加 time.Sleep。
	}

	// 所有重试均失败
	global.Logger.Error("All retries exhausted", "url", urlStr, "final_error", lastErr.Error())
	// 可以在这里记录失败列表到 struct 的另一个字段，如 this.Failed
}

// calculateDelay 计算当前应该休眠的时间
// 逻辑: BaseDelay + RandomRange + BackoffPenalty
func (this *Target) calculateDelay(cfg config.BaseConfig, attempt int) time.Duration {
	baseDelayMs := cfg.Delay
	randomRangeMs := cfg.RandomDelay
	backoffFactor := cfg.BackoffPolicy
	retryDelayMs := cfg.RetryDelay
	timeoutMs := cfg.Timeout

	// 1. 基础延迟 [0, baseDelay]
	delay := 0
	if baseDelayMs > 0 {
		delay = rand.Intn(baseDelayMs)
	}

	// 2. 随机延迟叠加 [0, randomRange]
	if randomRangeMs > 0 {
		delay += rand.Intn(randomRangeMs + 1)
	}

	// 3. 重试退避 (仅在 attempt > 0 时生效，即重试时)
	if attempt > 0 && backoffFactor >= 1 {
		// 确定基础等待时间：如果配置了 RetryDelay 则用之，否则用 Timeout
		waitBase := retryDelayMs
		if waitBase == 0 {
			waitBase = timeoutMs
		}

		// 公式: waitBase * (factor ^ (attempt - 1))
		// 注意：attempt 从 0 开始，第一次重试是 attempt=1, 指数为 0 (即 1 倍)
		multiplier := math.Pow(float64(backoffFactor), float64(attempt-1))
		backoffMs := float64(waitBase) * multiplier

		// 避免溢出或过大，可设置上限 (例如 1 小时)，这里暂不设硬上限
		delay += int(backoffMs)
	}

	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay) * time.Millisecond
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
