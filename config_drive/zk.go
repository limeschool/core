package config_drive

import (
	"bytes"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
	"strings"
	"time"
)

type zookeeper struct {
	client *zk.Conn
	wait   *uint64
	path   string
	tp     string
}

func NewZK(conf *Config) ConfigService {
	client, _, err := zk.Connect(strings.Split(conf.Host, ","), 10*time.Second)
	if err != nil {
		panic("zk config init err:" + err.Error())
	}
	return &zookeeper{client: client, path: conf.Path, tp: conf.Type}
}

func (z *zookeeper) Init() *viper.Viper {
	v := viper.New()
	//v := viper.NewWithOptions(viper.)
	v.SetConfigType(z.tp)
	if err := z.Get(v); err != nil {
		panic("zk config get err:" + err.Error())
	}
	go z.Watch(v)
	return v
}

func (z *zookeeper) Get(v *viper.Viper) error {
	data, _, err := z.client.Get(z.path)
	if err != nil {
		return err
	}
	if err = v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		return err
	}
	return nil
}

func (z *zookeeper) Watch(v *viper.Viper) {
	for {
		data, _, event, err := z.client.GetW(z.path)
		if err != nil {
			break
		}
		evt := <-event
		if evt.Type == zk.EventNodeDataChanged {
			if err = v.ReadConfig(bytes.NewBuffer(data)); err != nil {
				continue
			}
			if CallBack != nil {
				CallBack(v)
			}
		}
	}
}
