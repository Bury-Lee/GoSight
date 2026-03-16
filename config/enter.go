package config

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

const ConfigPath = "config.json"

type ConfigEnter struct {
	Log Log `json:"log"`
}

type WebConfig struct {
	Root          *WebConfig             `json:"root"`           //根配置
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

const UserDataPath = "./userdata/"

var Config ConfigEnter
var DefaultWebConfigs WebConfig

var WebConfigList map[string]*WebConfig = make(map[string]*WebConfig)
var nameList map[string]bool = make(map[string]bool)

func init() {
	// 注册所有可能通过 interface{} 传输的类型（如果有的话）
	// gob.Register(...)
}

func Load() error {
	return errors.New("TODO:读取自定义配置")
}

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

func GetWebConf(conf string) (error, *WebConfig) {
	if conf == "" {
		return errors.New("配置名不能为空,使用默认配置"), &DefaultWebConfigs
	}
	V, ok := WebConfigList[conf]
	if !ok || V == nil {
		return errors.New("配置不存在,使用默认配置"), &DefaultWebConfigs
	}
	return nil, V
}

func (c *ConfigEnter) Init(filepath string) error {
	conf, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	json.Unmarshal(conf, c)
	json.Unmarshal(conf, &DefaultWebConfigs)
	return nil
}

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

	// 处理 Root 绑定
	if webConf.RootName == "" {
		webConf.Root = &DefaultWebConfigs
		webConf.RootName = "default"
	} else {
		err, webConf.Root = GetWebConf(webConf.RootName)
		if err != nil {
			// 如果指定的 root 不存在，绑定到 default
			webConf.Root = &DefaultWebConfigs
			webConf.RootName = "default"
		}
	}

	// 处理 Next 绑定
	webConf.Next = make([]*WebConfig, 0, len(webConf.NextName))
	for _, name := range webConf.NextName {
		err, nextConf := GetWebConf(name)
		if err != nil {
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

// SaveAsGobWriter 将配置编码到 io.Writer（适用于网络传输或内存缓冲）
func (c *WebConfig) SaveAsGobWriter(w *bytes.Buffer) error {
	// 自动绑定 Root
	if c.Root == nil || c.RootName == "" {
		c.Root = &DefaultWebConfigs
		c.RootName = "default"
	}

	encoder := gob.NewEncoder(w)
	return encoder.Encode(c)
}

// LoadFromGobReader 从 io.Reader 解码配置
func LoadFromGobReader(r *bytes.Reader) (*WebConfig, error) {
	var webConf WebConfig
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&webConf); err != nil {
		return nil, fmt.Errorf("gob 解码失败: %w", err)
	}

	// 自动绑定 Root（同上）
	if webConf.Root == nil {
		if webConf.RootName == "" {
			webConf.Root = &DefaultWebConfigs
			webConf.RootName = "default"
		} else {
			err, rootConf := GetWebConf(webConf.RootName)
			if err != nil {
				webConf.Root = &DefaultWebConfigs
				webConf.RootName = "default"
			} else {
				webConf.Root = rootConf
			}
		}
	}

	return &webConf, nil
}
