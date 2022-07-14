package config_drive

import (
	"bytes"
	"context"
	"errors"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/api/v3/mvccpb"
	etcdApi "go.etcd.io/etcd/client/v3"
	"strings"
)

type etcd struct {
	client *etcdApi.Client
	wait   *uint64
	path   string
	tp     string
}

func NewEtcd(conf *Config) ConfigService {
	client, err := etcdApi.New(etcdApi.Config{
		Endpoints:   strings.Split(conf.Host, ","),
		DialTimeout: 10,
		Username:    conf.Username,
		Password:    conf.Password,
	})
	if err != nil {
		panic("etcd config init err:" + err.Error())
	}
	return &etcd{client: client, path: conf.Path, tp: conf.Type}
}

func (c *etcd) Init() *viper.Viper {
	v := viper.New()
	v.SetConfigType(c.tp)
	if err := c.Get(v); err != nil {
		panic("etcd config get err:" + err.Error())
	}
	go c.Watch(v)
	return v
}

// Get 从中间件中获取配置
func (c *etcd) Get(v *viper.Viper) error {
	data, err := c.client.KV.Get(context.TODO(), c.path)
	if err != nil || len(data.Kvs) == 0 {
		return errors.New("get kv is fail:" + err.Error())
	}
	if err = v.ReadConfig(bytes.NewBuffer(data.Kvs[0].Value)); err != nil {
		return err
	}
	return nil
}

// Watch 监听配置变更
func (c *etcd) Watch(v *viper.Viper) {
	for {
		ch := c.client.Watch(context.TODO(), c.path)
		for ws := range ch {
			for _, event := range ws.Events {
				switch event.Type {
				case mvccpb.PUT:
					if err := v.ReadConfig(bytes.NewBuffer(event.Kv.Value)); err != nil {
						continue
					}
					if CallBack != nil {
						CallBack(v)
					}
				}
			}
		}
	}
}
