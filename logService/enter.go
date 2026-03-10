// log/enter.go
// 日志模块
package logService

import (
	"GoSight/config"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"
)

var Logger *slog.Logger
var LogFile *os.File

func InitLog() { //TODO:改良日志格式
	Time := "2006-01-02"
	fileName := fmt.Sprintf("%s.log", time.Now().Format(Time))
	err := os.MkdirAll(config.Config.Log.LogPath, 0755)
	if err != nil {
		fmt.Printf("无法创建日志目录: %v\n", err)
		return
	}
	LogFile, err = os.OpenFile(path.Join(config.Config.Log.LogPath, fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("无法打开日志文件: %v\n", err)
		return
	}

	// 初始化日志
	opt := &slog.HandlerOptions{
		Level:     config.Config.Log.GetLogLevel(),
		AddSource: config.Config.Log.IsAddSource,
	}
	var handler slog.Handler
	multiWriter := io.MultiWriter(os.Stdout, LogFile)

	switch config.Config.Log.LogFormat {
	case "json":
		{
			handler = slog.NewJSONHandler(multiWriter, opt)
		}
	case "text":
		{
			handler = slog.NewTextHandler(multiWriter, opt)
		}
	default:
		{
			fmt.Printf("未知的日志格式: %s，默认使用 JSON 格式\n", config.Config.Log.LogFormat)
			handler = slog.NewJSONHandler(multiWriter, opt)
		}
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger) //设置默认服务器
}
