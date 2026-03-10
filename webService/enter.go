package webservice

import (
	"GoSight/config"
	"errors"
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
