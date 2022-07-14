package config_drive

import (
	"bytes"
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

type local struct {
	client *os.File
	wait   *uint64
	path   string
	tp     string
}

func NewLocal(conf *Config) ConfigService {
	client, err := os.Open(conf.Path)
	if err != nil {
		panic("local config init err:" + err.Error())
	}
	return &local{client: client, path: conf.Path, tp: conf.Type}
}

func (c *local) Init() *viper.Viper {
	v := viper.New()
	v.SetConfigType(c.tp)
	if err := c.Get(v); err != nil {
		panic("local config get err:" + err.Error())
	}
	go c.Watch(v)
	return v
}

// Get 从中间件中获取配置
func (c *local) Get(v *viper.Viper) error {
	c.client.Seek(0, 0) //重制到最开始
	data, err := ioutil.ReadAll(c.client)
	if err != nil {
		return errors.New("get local config fail:" + err.Error())
	}
	if err = v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		return err
	}
	return nil
}

func (c *local) Watch(v *viper.Viper) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	if err = watcher.Add(c.path); err != nil {
		panic(err)
	}

	defer watcher.Close()

	for {
		select {
		case event, ok := <-watcher.Events: // 正常的事件的处理逻辑
			if !ok {
				continue
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err = c.Get(v); err != nil {
					CallBack(v)
				}
			}
		case _, ok := <-watcher.Errors:
			if !ok {
				break
			}
		}
	}
}
