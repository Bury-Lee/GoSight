下一步就是写轮番爬取和多爬取了

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// 单个文件上限 10MB
	MaxFileSize = 10 * 1024 * 1024
	// 目录总上限 50MB（防止太多小文件）
	MaxTotalSize = 50 * 1024 * 1024
)

func AllLoad(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return loadFromDir(path)
	}
	return loadFromFile(path)
}

// 从单个文件加载
func loadFromFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > MaxFileSize {
		return fmt.Errorf("文件 %s 超过大小限制 %d bytes", path, MaxFileSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return parseConfig(data)
}

// 从目录加载，合并所有 JSON
package main

import (
    "fmt"
    "os"
)

func main() {
    file, err := os.Open("example.txt")
    if err != nil {
        fmt.Println("打开文件失败:", err)
        return
    }
    defer file.Close()

    buffer := make([]byte, 1024) // 1KB 缓冲区
    for {
        n, err := file.Read(buffer)
        if err != nil {
            if err.Error() == "EOF" {
                break // 文件结束
            }
            fmt.Println("读取错误:", err)
            return
        }
        fmt.Println(string(buffer[:n]))
    }
}
```