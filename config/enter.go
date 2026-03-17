package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// ConfigPath 默认配置文件路径
const ConfigPath = "config.json"

// ConfigEnter 主配置入口结构体
// 参数:Log - 日志配置项
// 说明:作为整个爬虫系统的配置入口，包含日志等基础配置
type ConfigEnter struct {
	Log Log `json:"log"`
}

// WebConfig Web配置结构体
// 参数:Root - 根配置指针,RootName - 根配置名称,ConfigName - 配置名称,Config - 基础配置
// 参数:Web - 爬取网站配置,Agents - 代理配置,CustomDomains - 自定义域名配置映射
// 参数:Render - 渲染配置,Next - 下一级配置列表,NextName - 下一级配置名称列表
// 说明:支持层级配置结构，可继承父配置，支持自定义域名和代理设置
type WebConfig struct {
	Root          *WebConfig             `json:"-"`              //根配置
	RootName      string                 `json:"root_name"`      //根配置名
	ConfigName    string                 `json:"config_name"`    //配置名
	Config        BaseConfig             `json:"base_config"`    //基础配置
	Web           BlackConfig            `json:"web"`            //爬取网站配置
	Agents        Agent                  `json:"agents"`         //代理配置
	CustomDomains map[string]*BaseConfig `json:"custom_domains"` //自定义域名配置
	Render        RenderConfig           `json:"render"`         //渲染配置
	Next          []*WebConfig           `json:"next"`           //下一级配置
	NextName      []string               `json:"next_name"`      //下一级配置名
}

// UserDataPath 用户数据存储路径
// 说明:用于存储用户配置文件，后续将迁移到global包中管理
const UserDataPath = "./userdata/"

// Config 全局主配置实例
// 说明:存储系统的主配置信息，全局共享使用
var Config ConfigEnter

// DefaultWebConfigs 默认Web配置实例
// 说明:作为所有Web配置的默认模板，提供基础配置项
var DefaultWebConfigs WebConfig

// WebConfigList Web配置名称映射表
// 说明:通过配置名称快速查找对应的Web配置实例，支持配置复用
var WebConfigList map[string]*WebConfig = make(map[string]*WebConfig) //名字列表

// nameList 配置名称存在性检查表
// 说明:用于快速检查配置名称是否已被使用，防止重复命名
var nameList map[string]bool = make(map[string]bool) //配置名是否存在

// AllLoad 批量加载配置文件
// 参数:path - 配置文件路径或目录路径
// 返回:错误列表 - 加载过程中遇到的所有错误
// 说明:支持单文件和目录批量加载，自动过滤超大文件(8MB限制)，递归处理目录下所有JSON文件
func AllLoad(path string) []error {
	info, err := os.Stat(path)
	if err != nil {
		return []error{fmt.Errorf("无法访问路径 %s: %w", path, err)}
	}

	var result []error

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return []error{fmt.Errorf("读取目录失败 %s: %w", path, err)}
		}

		for _, entry := range entries {
			// 跳过子目录，只处理文件
			if entry.IsDir() {
				continue
			}

			fileInfo, err := entry.Info()
			if err != nil {
				result = append(result, fmt.Errorf("获取文件信息失败 %s: %w", entry.Name(), err))
				continue
			}

			// 检查文件大小
			if fileInfo.Size() > 1024*1024*8 {
				result = append(result, fmt.Errorf("文件 %s 大小超过限制 (%d > %d bytes)", entry.Name(), fileInfo.Size(), 1024*1024*8))
				continue
			}

			// 构造完整路径并加载
			fullPath := filepath.Join(path, entry.Name())

			// 注意：原代码逻辑是直接调用 LoadFromJSON(v.Name())，这通常只在当前目录下有效。
			// 建议传入完整路径，或者确保工作目录正确。这里假设 LoadFromJSON 需要完整路径。
			if _, loadErr := LoadFromJSON(fullPath); loadErr != nil {
				result = append(result, fmt.Errorf("加载文件 %s 失败: %w", entry.Name(), loadErr))
			}
		}
	} else {
		// 处理单个文件
		if info.Size() > 1024*1024*8 {
			return []error{fmt.Errorf("文件 %s 大小超过限制 (%d > %d bytes)", path, info.Size(), 1024*1024*8)}
		}

		if _, loadErr := LoadFromJSON(path); loadErr != nil {
			return []error{fmt.Errorf("加载文件 %s 失败: %w", path, loadErr)}
		}
	}

	return result
}

// Add 添加Web配置到全局列表
// 参数:this - Web配置实例指针
// 返回:错误信息 - 配置名重复时返回重命名警告
// 说明:自动检测配置名是否已存在，存在则自动重命名并返回警告信息
func (this *WebConfig) Add() error {
	var err error
	if exit, _ := nameList[this.ConfigName]; exit {
		this.ConfigName = this.ConfigName + "(1)"
		err = errors.New("配置名已存在,自动起名为: " + this.ConfigName)
	}

	if this == nil {
		return errors.New("配置不能为空")
	}
	// 实际应该在这里加入列表
	WebConfigList[this.ConfigName] = this
	nameList[this.ConfigName] = true
	return err
}

// Add 添加指定名称的Web配置到全局列表
// 参数:name - 配置名称,webConfig - Web配置实例指针
// 返回:错误信息 - 配置名重复时返回重命名警告
// 说明:支持自定义配置名称，自动处理名称冲突并注册到全局配置列表
func Add(name string, webConfig *WebConfig) error {
	var err error
	if exit, _ := nameList[name]; exit {
		name = name + "(1)"
		err = errors.New("配置名已存在,自动起名为: " + name)
	}

	if webConfig == nil {
		return errors.New("配置不能为空")
	}
	// 实际应该在这里加入列表
	WebConfigList[name] = webConfig
	nameList[name] = true
	return err
}

// GetWebConf 根据配置名称获取Web配置
// 参数:conf - 配置名称
// 返回:错误信息 - 配置不存在或名称为空时返回错误,WebConfig - 对应的配置实例(不存在时返回默认配置)
// 说明:提供安全的配置获取方式，配置不存在时变成nil
func GetWebConf(conf string) (error, *WebConfig) {
	if conf == "" {
		return errors.New("配置名不能为空"), nil
	}
	V, ok := WebConfigList[conf]
	if !ok || V == nil {
		return errors.New("配置不存在"), nil
	}
	return nil, V
}

// Init 初始化主配置
// 参数:filepath - 配置文件路径
// 返回:错误信息 - 文件读取或解析失败时返回错误
// 说明:从JSON文件读取配置并解析到全局Config和DefaultWebConfigs中
func (c *ConfigEnter) Init(filepath string) error {
	conf, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	json.Unmarshal(conf, c)                  //给基本配置赋值
	json.Unmarshal(conf, &DefaultWebConfigs) //给默认配置赋值
	return nil
}

// SaveAsJson 将Web配置保存为JSON文件
// 参数:c - Web配置实例指针
// 返回:错误信息 - 序列化或文件写入失败时返回错误
// 说明:自动在用户数据目录下创建以配置名命名的JSON文件，格式化输出便于阅读
func (c *WebConfig) SaveAsJson() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	err = os.WriteFile(path.Join(UserDataPath, c.ConfigName+".json"), data, 0644)
	if err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}
	return nil
}

// LoadFromJSON 从JSON文件读取并绑定Web配置
// 参数:filename - JSON配置文件路径
// 返回:WebConfig - 解析后的配置实例,错误信息 - 文件读取或解析失败时返回错误
// 说明:自动处理配置层级绑定，包括Root父配置和Next子配置的关联，注册到全局配置列表
// 不过不会递归处理
// 1. 返回指针是为了确保调用者获取到的是注册到全局列表中的同一实例（引用语义），避免数据不一致。
// 2. 避免大结构体拷贝，提升性能。
// 3. 自动处理配置层级绑定（Root父配置和Next子配置关联）。
func LoadFromJSON(filename string) (*WebConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}

	var webConf WebConfig
	err = json.Unmarshal(data, &webConf)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	// 1. 确定 Root 节点
	if webConf.RootName == "" {
		// 情况 A: 未指定 RootName，直接使用默认配置
		webConf.Root = &DefaultWebConfigs
		webConf.RootName = "default"
	} else {
		// 情况 B: 指定了 RootName，尝试获取
		err, root := GetWebConf(webConf.RootName)
		if err != nil || root == nil {
			// 情况 B-1: 获取失败或不存在，降级为默认配置
			webConf.Root = &DefaultWebConfigs
			webConf.RootName = "default"
			// 可选：这里可以加一行日志记录降级行为，但不强制要求
		} else {
			// 情况 B-2: 获取成功
			webConf.Root = root
		}
	}
	// 2. 统一执行绑定逻辑 (将当前 webConf 挂载到确定的 Root 下)
	// 此时 webConf.Root 必然不为 nil (要么是查找到的，要么是 DefaultWebConfigs)
	webConf.Root.Next = append(webConf.Root.Next, &webConf)
	webConf.Root.NextName = append(webConf.Root.NextName, webConf.ConfigName)

	// 处理 Next 绑定
	webConf.Next = make([]*WebConfig, 0, len(webConf.NextName))
	for _, name := range webConf.NextName {
		err, nextConf := GetWebConf(name)
		if err != nil || nextConf == nil {
			continue // 跳过不存在的子配置
		}
		webConf.Next = append(webConf.Next, nextConf)
	}

	// 注册到全局列表
	if webConf.ConfigName != "" {
		WebConfigList[webConf.ConfigName] = &webConf
		nameList[webConf.ConfigName] = true
	}

	return &webConf, nil
}
