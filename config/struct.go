package config

import (
	"log/slog"
	"net/http"
)

type BaseConfig struct { //配置项
	Output        string `json:"output"`         //输出文件路径
	Concurrency   int    `json:"concurrency"`    //并发数
	Delay         int    `json:"delay"`          //延迟时间,启用随机延迟时,每个请求的延迟时间范围为[0,delay]毫秒
	RandomDelay   int    `json:"random_delay"`   //随机延迟时间范围（毫秒）
	Timeout       int    `json:"timeout"`        //超时时间
	MaxRetries    int    `json:"max_retries"`    //最大重试次数
	RetryDelay    int    `json:"retry_delay"`    //重试延迟时间（毫秒）,为0时与超时时间相同
	BackoffPolicy int    `json:"backoff_policy"` //重试退避因子,重试等待时间计算公式通常为：delay * (backoff_factor ^ (retry_count - 1))。<1时不启用
	Max_log_size  int    `json:"max_log_size"`   //日志文件最大大小（MB）,避免极端情况发生
}

type RenderConfig struct {
	Enable    bool   `json:"enable"`     // 是否启用浏览器渲染
	Engine    string `json:"engine"`     // "chromium", "firefox", "webkit"
	Headless  bool   `json:"headless"`   // 是否启用无头模式
	WaitUntil string `json:"wait_until"` // "load", "domcontentloaded", "networkidle"
}

type BlackConfig struct {
	BlackList []string `json:"black_list"` //黑名单,包含IP或域名,格式为"IP/域名"的前缀
}

type Agent struct {
	Proxy                 string        `json:"proxy"`                   //代理地址
	ProxyRotationStrategy string        `json:"proxy_rotation_strategy"` //代理轮换策略
	ID                    string        `json:"id"`                      //代理ID
	IP                    string        `json:"ip"`                      //代理IP
	Headers               http.Header   `json:"headers"`                 //请求头
	Cookies               []http.Cookie `json:"cookies"`                 //请求头
}

type Log struct {
	LogFormat   string `json:"log_format"`    //日志格式
	LogLevel    string `json:"log_level"`     //日志级别
	LogPath     string `json:"log_path"`      //日志文件路径
	IsAddSource bool   `json:"is_add_source"` //是否添加源信息,默认false
}

func (Log *Log) GetLogLevel() slog.Level {
	switch Log.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
