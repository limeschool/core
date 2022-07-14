package config_drive

import (
	"bytes"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	"log"
)

type consul struct {
	client *consulApi.Client
	wait   *uint64
	path   string
	tp     string
}

// NewConsul 创建consul对象，用于获取和监听日志
func NewConsul(conf *Config) ConfigService {
	client, err := consulApi.NewClient(&consulApi.Config{
		Address: conf.Host,
		Token:   conf.Password,
	})
	if err != nil {
		panic("consul config init err:" + err.Error())
	}
	return &consul{client: client, path: conf.Path, tp: conf.Type}
}

// Init 初始化consul配置信息
func (c *consul) Init() *viper.Viper {
	v := viper.New()
	v.SetConfigType(c.tp)
	if err := c.Get(v); err != nil {
		panic("consul config get err:" + err.Error())
	}
	go c.Watch(v)
	return v
}

// Get 从中间件中获取配置
func (c *consul) Get(v *viper.Viper) error {
	var q *consulApi.QueryOptions
	if c.wait != nil {
		q = &consulApi.QueryOptions{
			WaitIndex: *c.wait,
		}
	}
	data, meta, err := c.client.KV().Get(c.path, q)
	if err != nil {
		return err
	}
	c.wait = &meta.LastIndex
	if err = v.ReadConfig(bytes.NewBuffer(data.Value)); err != nil {
		return err
	}
	return nil
}

// Watch 监听配置变更
func (c *consul) Watch(v *viper.Viper) {
	for {
		if err := c.Get(v); err != nil {
			log.Println("consul监听变更信息获取失败")
			continue
		}
		if CallBack != nil {
			CallBack(v)
		}
	}
}
